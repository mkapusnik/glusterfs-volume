package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
)

const defaultMode = 0o755

type activeMount struct {
	connections int
	mountpoint  string
	createdAt   time.Time
	ids         map[string]int
}

type glusterfsDriver struct {
	sync.RWMutex

	root           string
	store          *stateStore
	volumes        map[string]volumeState
	mounts         map[string]*activeMount
	defaultVolume  string
	defaultServers []string
	client         glfsConnector
}

func (d *glusterfsDriver) Create(r *volume.CreateRequest) error {
	d.Lock()
	defer d.Unlock()

	servers := splitList(r.Options["server"])
	name := r.Options["name"]
	if name == "" {
		name = r.Options["volume"]
	}
	if len(servers) == 0 {
		servers = d.defaultServers
	}
	if name == "" {
		name = d.defaultVolume
	}
	volumeName, subdir, err := parseGfsName(name)
	if err != nil {
		return err
	}
	if volumeName == "" || len(servers) == 0 {
		return fmt.Errorf("glusterfs options must include server and volume")
	}

	if _, ok := d.volumes[r.Name]; !ok {
		d.volumes[r.Name] = volumeState{
			Name:      r.Name,
			Servers:   servers,
			Volume:    volumeName,
			Subdir:    subdir,
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		}
		if err := d.store.save(d.volumes); err != nil {
			return err
		}
	}

	return nil
}

func (d *glusterfsDriver) List() (*volume.ListResponse, error) {
	d.RLock()
	defer d.RUnlock()

	vols := make([]*volume.Volume, 0, len(d.volumes))
	for _, v := range d.volumes {
		vols = append(vols, &volume.Volume{Name: v.Name})
	}

	return &volume.ListResponse{Volumes: vols}, nil
}

func (d *glusterfsDriver) Get(r *volume.GetRequest) (*volume.GetResponse, error) {
	d.RLock()
	defer d.RUnlock()

	state, ok := d.volumes[r.Name]
	if !ok {
		return &volume.GetResponse{}, fmt.Errorf("volume %s not found", r.Name)
	}

	status := map[string]interface{}{
		"servers": state.Servers,
		"volume":  state.Volume,
		"subdir":  state.Subdir,
	}
	if mount, ok := d.mounts[r.Name]; ok {
		status["mountpoint"] = mount.mountpoint
	}

	vol := &volume.Volume{
		Name:       state.Name,
		CreatedAt:  state.CreatedAt,
		Mountpoint: state.Subdir,
		Status:     status,
	}

	return &volume.GetResponse{Volume: vol}, nil
}

func (d *glusterfsDriver) Remove(r *volume.RemoveRequest) error {
	d.Lock()
	defer d.Unlock()

	if mount, ok := d.mounts[r.Name]; ok && mount.connections > 0 {
		return fmt.Errorf("volume %s is still mounted", r.Name)
	}

	delete(d.volumes, r.Name)
	delete(d.mounts, r.Name)

	return d.store.save(d.volumes)
}

func (d *glusterfsDriver) Path(r *volume.PathRequest) (*volume.PathResponse, error) {
	d.RLock()
	defer d.RUnlock()

	mount, ok := d.mounts[r.Name]
	if !ok || mount.connections == 0 {
		return &volume.PathResponse{}, fmt.Errorf("no mountpoint for volume")
	}

	return &volume.PathResponse{Mountpoint: mount.mountpoint}, nil
}

func (d *glusterfsDriver) Mount(r *volume.MountRequest) (*volume.MountResponse, error) {
	d.Lock()
	defer d.Unlock()

	state, ok := d.volumes[r.Name]
	if !ok {
		return &volume.MountResponse{}, fmt.Errorf("volume %s not found", r.Name)
	}

	mountpoint := d.mountpoint(r.Name)
	info, ok := d.mounts[r.Name]
	if !ok {
		info = &activeMount{mountpoint: mountpoint, ids: map[string]int{}, createdAt: time.Now().UTC()}
		d.mounts[r.Name] = info
	}

	stat, err := os.Lstat(mountpoint)
	if err != nil || info.connections == 0 {
		if err != nil && !os.IsNotExist(err) {
			_ = d.client.unmount(mountpoint)
		}
		if os.IsNotExist(err) {
			if err := os.MkdirAll(mountpoint, defaultMode); err != nil && !os.IsExist(err) {
				return &volume.MountResponse{}, err
			}
		}
		stat, err = os.Lstat(mountpoint)
		if err != nil {
			return &volume.MountResponse{}, err
		}
		if !stat.IsDir() {
			if err := os.Remove(mountpoint); err != nil {
				return &volume.MountResponse{}, err
			}
			if err := os.MkdirAll(mountpoint, defaultMode); err != nil {
				return &volume.MountResponse{}, err
			}
		}

		if state.Subdir != "" {
			if err := d.client.mountWithGlusterfs(mountpoint, state.Volume, state.Servers, ""); err != nil {
				return &volume.MountResponse{}, err
			}
			if err := os.MkdirAll(filepath.Join(mountpoint, state.Subdir), defaultMode); err != nil {
				_ = d.client.unmount(mountpoint)
				return &volume.MountResponse{}, err
			}
			if err := d.client.unmount(mountpoint); err != nil {
				return &volume.MountResponse{}, err
			}
		}

		if err := d.client.mountWithGlusterfs(mountpoint, state.Volume, state.Servers, state.Subdir); err != nil {
			return &volume.MountResponse{}, err
		}
	}

	info.mountpoint = mountpoint
	info.ids[r.ID]++
	info.connections++

	return &volume.MountResponse{Mountpoint: mountpoint}, nil
}

func (d *glusterfsDriver) Unmount(r *volume.UnmountRequest) error {
	d.Lock()
	defer d.Unlock()

	info, ok := d.mounts[r.Name]
	if !ok {
		return fmt.Errorf("volume not mounted: %s", r.Name)
	}
	if info.connections == 0 {
		return fmt.Errorf("volume has no active mounts: %s", r.Name)
	}
	count, ok := info.ids[r.ID]
	if !ok {
		return fmt.Errorf("mount %s does not know about client %s", r.Name, r.ID)
	}

	count--
	info.connections--
	if count <= 0 {
		delete(info.ids, r.ID)
	} else {
		info.ids[r.ID] = count
	}

	if len(info.ids) == 0 {
		log.Printf("Unmounting volume %s", r.Name)
		if err := d.client.unmount(info.mountpoint); err != nil {
			return err
		}
		delete(d.mounts, r.Name)
	}

	return nil
}

func (d *glusterfsDriver) Capabilities() *volume.CapabilitiesResponse {
	return &volume.CapabilitiesResponse{Capabilities: volume.Capability{Scope: "global"}}
}

func (d *glusterfsDriver) mountpoint(name string) string {
	return filepath.Join(d.root, name)
}

package main

import (
	"log"
	"os"
	"strings"

	"github.com/docker/go-plugins-helpers/volume"
)

//------------------------------

// config.json settings
const socketAddress = "glusterfs"
const propagatedMount = "/mnt/volumes"

// -------------
// main

func init() {
	log.SetFlags(0)
	logfile := os.Getenv("LOGFILE")
	if logfile != "" {
		f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}
}

func main() {
	gfsvol := os.Getenv("GFS_VOLUME")
	gfsservers := strings.Split(os.Getenv("GFS_SERVERS"), ",")

	d := &glusterfsDriver{
		mounts: map[string]*activeMount{},
		root:   propagatedMount,
		client: glfsConnector{
			conn:   nil,
			volume: gfsvol,
			hosts:  gfsservers,
		},
	}

	h := volume.NewHandler(d)
	log.Printf("GlusterFS Volume Plugin listening on %s.sock", socketAddress)
	err := h.ServeUnix(socketAddress, 0)
	log.Print(err)
	return
}

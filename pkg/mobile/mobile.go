package mobile

import (
	"fileshare/pkg/discovery"
	"fileshare/pkg/server"
)

var srv *server.Server
var disc *discovery.Service

// StartServer initializes the background Go server for the Android app
func StartServer(port int, downloadDir string) error {
	reg := discovery.NewPeerRegistry()
	var err error
	disc, err = discovery.NewService(port, reg)
	if err != nil {
		return err
	}
	disc.Start()
	
	srv = server.NewServer(port, downloadDir, disc, true, "")
	go srv.Start()
	return nil
}

// StopServer cleanly tears down the Go server
func StopServer() {
	if srv != nil {
		srv.Stop()
	}
	if disc != nil {
		disc.Stop()
	}
}

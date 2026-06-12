package main

import (
	"storage-management/internal/server"
)

const (
	IP_ADDR = "0.0.0.0:6969"
)

func main() {
	server := server.NewServer(IP_ADDR)
	server.Routes()
	server.Run()
}

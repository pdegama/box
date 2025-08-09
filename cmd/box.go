package main

import (
	"fmt"
	"os"

	clientapp "github.com/pdegama/box/client"
	"github.com/pdegama/box/config"
	serverapp "github.com/pdegama/box/server"
)

func main() {
	mode := os.Getenv("MODE")
	verbose := os.Getenv("VERBOSE")

	config.LoadConfig()

	if verbose == "true" {
		fmt.Println(config.ConfOpts)
	}

	if mode != "client" {
		serverapp.StartServer()
	} else {
		clientapp.StartClient()
	}
}

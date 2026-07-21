package main

import (
	"fmt"
	"os"
	"sync"

	clientapp "github.com/rellitelink/box/client"
	"github.com/rellitelink/box/config"
	serverapp "github.com/rellitelink/box/server"
)

func main() {
	verbose := os.Getenv("VERBOSE")

	config.LoadConfig()

	scWg := sync.WaitGroup{}
	scWg.Add(2)

	if verbose == "true" {
		fmt.Println(config.ConfOpts)
	}

	go func() {
		defer scWg.Done()
		serverapp.StartServer()
	}()

	go func() {
		defer scWg.Done()
		clientapp.StartClient()
	}()

	scWg.Wait()
}

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/polevych/in-memory-database/server"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	srv := server.NewServer()

	select {
		case <-stop: {
			srv.Stop()
		}
	}

}
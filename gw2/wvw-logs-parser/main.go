package main

import (
	"log"

	"github.com/ryan/wvwlog/server"
)

func main() {
	log.Println("Starting WvW Log Analyzer Dashboard")

	server.LoadFights()
	server.StartWatcher()

	if err := server.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}

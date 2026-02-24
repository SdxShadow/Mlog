package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mlog/mlog/internal/config"
	"github.com/mlog/mlog/internal/db"
)

var (
	configPath = flag.String("config", "/etc/mlog/mlog.yaml", "Path to config file")
	version    = flag.Bool("version", false, "Show version")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("mlogd version 0.1.0")
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting mlogd on server: %s", cfg.Server.ID)

	if err := db.Init(cfg.Database.Path); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutting down mlogd...")
}

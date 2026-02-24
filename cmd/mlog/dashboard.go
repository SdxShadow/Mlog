package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mlog/mlog/internal/db"
	"github.com/mlog/mlog/pkg/types"
)

type Dashboard struct {
	events    []*types.Event
	mu        sync.RWMutex
	stopCh    chan bool
	maxLines  int
}

func NewDashboard() *Dashboard {
	return &Dashboard{
		events:   make([]*types.Event, 0),
		stopCh:   make(chan bool),
		maxLines: 50,
	}
}

func (d *Dashboard) Start() error {
	if err := initDB(); err != nil {
		return err
	}
	defer db.Close()

	go d.pollEvents()

	d.render()
	return nil
}

func (d *Dashboard) pollEvents() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			events, err := db.QueryEvents(&db.EventQuery{Limit: 20})
			if err != nil {
				continue
			}

			d.mu.Lock()
			d.events = events
			d.mu.Unlock()

			d.render()
		case <-d.stopCh:
			return
		}
	}
}

func (d *Dashboard) render() {
	d.mu.RLock()
	events := d.events
	d.mu.RUnlock()

	clearScreen()
	
	d.printHeader()
	d.printStats()
	d.printLogs(events)
	d.printFooter()
}

func (d *Dashboard) printHeader() {
	fmt.Print("\033[1;34m")
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                              MLOG MONITOR                                            │")
	fmt.Println("└──────────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Print("\033[0m")
}

func (d *Dashboard) printStats() {
	fmt.Print("\033[1;33m")
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ SYSTEM STATUS                                                                    │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────────────────┤")

	cmd := exec.Command("sh", "-c", "top -bn1 | head -15")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.NewFile(0, os.DevNull)
	cmd.Run()

	fmt.Println("└──────────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Print("\033[0m")
}

func (d *Dashboard) printLogs(events []*types.Event) {
	fmt.Print("\033[1;36m")
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────────────────┐")
	fmt.Println("│ LIVE LOGS                                                                        │")
	fmt.Println("├──────────────────────────────────────────────────────────────────────────────────────┤")
	fmt.Print("\033[0m")

	for _, e := range events {
		color := getColorForType(e.EventType)
		fmt.Print(color)
		
		ts := e.Timestamp.Format("15:04:05")
		shortType := strings.TrimPrefix(string(e.EventType), "SSH_")
		shortType = strings.TrimPrefix(shortType, "NGINX_")
		shortType = strings.TrimPrefix(shortType, "APACHE_")
		shortType = strings.TrimPrefix(shortType, "PM2_")
		
		if len(e.Message) > 50 {
			e.Message = e.Message[:47] + "..."
		}
		
		fmt.Printf("[\033[90m%s\033[0m] [\033[1m%8s\033[0m] %-20s %s\n", 
			ts, shortType, e.SourceIP, e.Message)
	}
	
	if len(events) == 0 {
		fmt.Println("\033[90m  Waiting for logs...\033[0m")
	}
	
	fmt.Print("\033[1;36m")
	fmt.Println("└──────────────────────────────────────────────────────────────────────────────────────┘")
	fmt.Print("\033[0m")
}

func (d *Dashboard) printFooter() {
	fmt.Println("\033[90m  Press Ctrl+C to exit\033[0m")
}

func getColorForType(t types.EventType) string {
	switch {
	case strings.Contains(string(t), "SSH_CONNECTED"):
		return "\033[32m"
	case strings.Contains(string(t), "SSH_FAILED"):
		return "\033[31m"
	case strings.Contains(string(t), "ERROR"):
		return "\033[31m"
	case strings.Contains(string(t), "WARNING"):
		return "\033[33m"
	case strings.Contains(string(t), "NGINX"):
		return "\033[35m"
	case strings.Contains(string(t), "APACHE"):
		return "\033[36m"
	case strings.Contains(string(t), "PM2"):
		return "\033[34m"
	default:
		return "\033[37m"
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

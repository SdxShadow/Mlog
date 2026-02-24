package monitor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/SdxShadow/Mlog/internal/db"
	"github.com/SdxShadow/Mlog/internal/parser/application"
	"github.com/SdxShadow/Mlog/internal/parser/ssh"
	"github.com/SdxShadow/Mlog/pkg/types"
)

type Watcher struct {
	serverID   string
	sshParser  *ssh.Parser
	nginxParser *application.NginxParser
	apacheParser *application.ApacheParser
	pm2Parser   *application.PM2Parser
	watcher    *fsnotify.Watcher
	files      map[string]int64
	stopCh     chan bool
}

func NewWatcher(serverID string) *Watcher {
	return &Watcher{
		serverID:    serverID,
		sshParser:   ssh.New(serverID),
		nginxParser: application.NewNginxParser(serverID),
		apacheParser: application.NewApacheParser(serverID),
		pm2Parser:   application.NewPM2Parser(serverID),
		files:       make(map[string]int64),
		stopCh:      make(chan bool),
	}
}

func (w *Watcher) AddPath(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Watch path does not exist: %s", path)
			return nil
		}
		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", path)
	}

	w.files[path] = stat.Size()
	return nil
}

func (w *Watcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = watcher

	for path := range w.files {
		if err := watcher.Add(path); err != nil {
			log.Printf("Failed to watch %s: %v", path, err)
		}
	}

	go w.run()
	return nil
}

func (w *Watcher) run() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				w.readNewLines(event.Name)
			}
		case <-w.stopCh:
			return
		}
	}
}

func (w *Watcher) readNewLines(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open %s: %v", path, err)
		return
	}
	defer file.Close()

	lastPos := w.files[path]
	file.Seek(lastPos, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		event := w.parseLine(path, line)
		if event != nil {
			if err := db.InsertEvent(event); err != nil {
				log.Printf("Failed to insert event: %v", err)
			}
		}
	}

	info, err := file.Stat()
	if err == nil {
		w.files[path] = info.Size()
	}
}

func (w *Watcher) parseLine(path, line string) *types.Event {
	ts := time.Now()

	if isAuthLog(path) {
		return w.sshParser.Parse(line, ts)
	}

	if isNginxAccess(path) {
		return w.nginxParser.ParseAccess(line, ts)
	}

	if isNginxError(path) {
		return w.nginxParser.ParseError(line, ts)
	}

	if isApacheAccess(path) {
		return w.apacheParser.ParseAccess(line, ts)
	}

	if isApacheError(path) {
		return w.apacheParser.ParseError(line, ts)
	}

	if isPM2Log(path) {
		return w.pm2Parser.Parse(line, ts)
	}

	return nil
}

func isAuthLog(path string) bool {
	return contains(path, "/var/log/auth.log", "/var/log/secure")
}

func isNginxAccess(path string) bool {
	return contains(path, "/nginx/access.log")
}

func isNginxError(path string) bool {
	return contains(path, "/nginx/error.log")
}

func isApacheAccess(path string) bool {
	return contains(path, "/apache2/access.log", "/httpd/access_log")
}

func isApacheError(path string) bool {
	return contains(path, "/apache2/error.log", "/httpd/error_log")
}

func isPM2Log(path string) bool {
	return contains(path, "/.pm2/logs/", "pm2.log")
}

func contains(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(s) >= len(sub) && (s[len(s)-len(sub):] == sub || s == sub) {
			return true
		}
		if len(s) >= len(sub) && indexOf(sub, s) >= 0 {
			return true
		}
	}
	return false
}

func indexOf(sub, s string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func (w *Watcher) Stop() error {
	w.stopCh <- true
	if w.watcher != nil {
		return w.watcher.Close()
	}
	return nil
}

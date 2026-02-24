package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/SdxShadow/Mlog/pkg/types"
)

var db *sql.DB

func Init(path string) error {
	// Create directory if not exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite3", path+"?_journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err = createSchema(); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT NOT NULL,
		server_id TEXT NOT NULL,
		event_type TEXT NOT NULL,
		severity TEXT NOT NULL,
		source_ip TEXT,
		dest_ip TEXT,
		source_port INTEGER,
		username TEXT,
		message TEXT,
		raw_log TEXT,
		metadata TEXT,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_events_event_type ON events(event_type);
	CREATE INDEX IF NOT EXISTS idx_events_source_ip ON events(source_ip);
	CREATE INDEX IF NOT EXISTS idx_events_username ON events(username);
	CREATE INDEX IF NOT EXISTS idx_events_severity ON events(severity);

	CREATE TABLE IF NOT EXISTS ssh_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		source_ip TEXT NOT NULL,
		source_port INTEGER,
		connected_at TEXT NOT NULL,
		disconnected_at TEXT,
		duration_seconds INTEGER,
		auth_method TEXT,
		client_version TEXT,
		status TEXT DEFAULT 'active'
	);

	CREATE TABLE IF NOT EXISTS security_incidents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		incident_type TEXT NOT NULL,
		severity TEXT NOT NULL,
		source_ip TEXT,
		start_time TEXT NOT NULL,
		end_time TEXT,
		event_count INTEGER DEFAULT 1,
		description TEXT,
		resolved INTEGER DEFAULT 0,
		metadata TEXT
	);

	CREATE TABLE IF NOT EXISTS config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS server_info (
		id TEXT PRIMARY KEY,
		hostname TEXT,
		os_version TEXT,
		arch TEXT,
		first_seen TEXT,
		last_seen TEXT
	);
	`

	_, err := db.Exec(schema)
	return err
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func InsertEvent(event *types.Event) error {
	query := `INSERT INTO events (timestamp, server_id, event_type, severity, source_ip, dest_ip, source_port, username, message, raw_log, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(query,
		event.Timestamp.Format(time.RFC3339),
		event.ServerID,
		event.EventType,
		event.Severity,
		event.SourceIP,
		event.DestIP,
		event.SourcePort,
		event.Username,
		event.Message,
		event.RawLog,
		event.MetadataJSON(),
	)

	return err
}

type EventQuery struct {
	EventType  string
	SourceIP   string
	Username   string
	Severity   string
	Since      *time.Time
	Until      *time.Time
	Limit      int
	Offset     int
}

func QueryEvents(q *EventQuery) ([]*types.Event, error) {
	query := "SELECT id, timestamp, server_id, event_type, severity, source_ip, dest_ip, source_port, username, message, raw_log, metadata FROM events WHERE 1=1"
	args := []interface{}{}

	if q.EventType != "" {
		query += " AND event_type LIKE ?"
		args = append(args, q.EventType+"%")
	}
	if q.SourceIP != "" {
		query += " AND source_ip = ?"
		args = append(args, q.SourceIP)
	}
	if q.Username != "" {
		query += " AND username = ?"
		args = append(args, q.Username)
	}
	if q.Severity != "" {
		query += " AND severity = ?"
		args = append(args, q.Severity)
	}
	if q.Since != nil {
		query += " AND timestamp >= ?"
		args = append(args, q.Since.Format(time.RFC3339))
	}
	if q.Until != nil {
		query += " AND timestamp <= ?"
		args = append(args, q.Until.Format(time.RFC3339))
	}

	query += " ORDER BY timestamp DESC"

	if q.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, q.Limit)
	} else {
		query += " LIMIT 100"
	}

	if q.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, q.Offset)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*types.Event
	for rows.Next() {
		e := &types.Event{}
		var timestamp, metadata string
		err := rows.Scan(&e.ID, &timestamp, &e.ServerID, &e.EventType, &e.Severity, &e.SourceIP, &e.DestIP, &e.SourcePort, &e.Username, &e.Message, &e.RawLog, &metadata)
		if err != nil {
			return nil, err
		}
		e.Timestamp, _ = time.Parse(time.RFC3339, timestamp)
		events = append(events, e)
	}

	return events, nil
}

func GetDB() *sql.DB {
	return db
}

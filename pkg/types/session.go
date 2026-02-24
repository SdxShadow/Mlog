package types

import (
	"net"
	"time"
)

type SSHSession struct {
	ID            int64     `json:"id"`
	SessionID     string    `json:"session_id"`
	Username      string    `json:"username"`
	SourceIP      string    `json:"source_ip"`
	SourcePort    int       `json:"source_port"`
	ConnectedAt   time.Time `json:"connected_at"`
	DisconnectedAt time.Time `json:"disconnected_at,omitempty"`
	Duration      int64     `json:"duration_seconds,omitempty"`
	AuthMethod    string    `json:"auth_method"`
	ClientVersion string    `json:"client_version,omitempty"`
	Status        string    `json:"status"`
}

func (s *SSHSession) SourceIPNet() net.IP {
	return net.ParseIP(s.SourceIP)
}

type SecurityIncident struct {
	ID           int64     `json:"id"`
	IncidentType string    `json:"incident_type"`
	Severity     Severity  `json:"severity"`
	SourceIP     string    `json:"source_ip,omitempty"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time,omitempty"`
	EventCount   int       `json:"event_count"`
	Description  string    `json:"description"`
	Resolved     bool      `json:"resolved"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

func (i *SecurityIncident) SetMetadata(key string, value interface{}) {
	if i.Metadata == nil {
		i.Metadata = make(map[string]interface{})
	}
	i.Metadata[key] = value
}

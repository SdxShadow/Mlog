package types

import (
	"encoding/json"
	"net"
	"time"
)

type EventType string

const (
	EventSSHConnected     EventType = "SSH_CONNECTED"
	EventSSHDisconnected EventType = "SSH_DISCONNECTED"
	EventSSHFailedAuth   EventType = "SSH_FAILED_AUTH"

	EventBruteForceSuspected EventType = "BRUTE_FORCE_SUSPECTED"
	EventPortScanSuspected   EventType = "PORT_SCAN_SUSPECTED"
	EventSudoSuccess        EventType = "SUDO_SUCCESS"
	EventSudoFailed         EventType = "SUDO_FAILED"

	EventServiceStarted EventType = "SERVICE_STARTED"
	EventServiceStopped EventType = "SERVICE_STOPPED"

	EventNginxRequest EventType = "NGINX_REQUEST"
	EventNginxError   EventType = "NGINX_ERROR"

	EventApacheRequest EventType = "APACHE_REQUEST"
	EventApacheError   EventType = "APACHE_ERROR"

	EventPM2Start    EventType = "PM2_START"
	EventPM2Stop     EventType = "PM2_STOP"
	EventPM2Restart  EventType = "PM2_RESTART"
	EventPM2Error    EventType = "PM2_ERROR"
	EventPM2Crash    EventType = "PM2_CRASH"

	EventCustom EventType = "CUSTOM"
)

type Severity string

const (
	SeverityDebug   Severity = "debug"
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
	SeverityCritical Severity = "critical"
)

type Event struct {
	ID            int64                  `json:"id"`
	Timestamp     time.Time              `json:"timestamp"`
	ServerID      string                 `json:"server_id"`
	EventType     EventType              `json:"event_type"`
	Severity      Severity               `json:"severity"`
	SourceIP      string                 `json:"source_ip,omitempty"`
	DestIP        string                 `json:"dest_ip,omitempty"`
	SourcePort    int                    `json:"source_port,omitempty"`
	Username      string                 `json:"username,omitempty"`
	Message       string                 `json:"message"`
	RawLog        string                 `json:"raw_log"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

func (e *Event) SetMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

func (e *Event) GetMetadata(key string) interface{} {
	if e.Metadata == nil {
		return nil
	}
	return e.Metadata[key]
}

func (e *Event) MetadataJSON() string {
	if e.Metadata == nil {
		return "{}"
	}
	b, _ := json.Marshal(e.Metadata)
	return string(b)
}

func ParseIP(ipStr string) net.IP {
	if ipStr == "" {
		return nil
	}
	return net.ParseIP(ipStr)
}

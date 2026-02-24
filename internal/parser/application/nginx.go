package application

import (
	"regexp"
	"strconv"
	"time"

	"github.com/mlog/mlog/pkg/types"
)

type NginxParser struct {
	serverID string
}

func NewNginxParser(serverID string) *NginxParser {
	return &NginxParser{serverID: serverID}
}

var nginxAccessPattern = regexp.MustCompile(`^(\S+)\s+-\s+(\S+)\s+\[([^\]]+)\]\s+"(\S+)\s+(\S+)\s+\S+"\s+(\d+)\s+(\d+)\s+"([^"]*)"\s+"([^"]*)"`)

func (p *NginxParser) ParseAccess(line string, ts time.Time) *types.Event {
	m := nginxAccessPattern.FindStringSubmatch(line)
	if len(m) < 10 {
		return nil
	}

	status, _ := strconv.Atoi(m[6])
	severity := types.SeverityInfo
	if status >= 500 {
		severity = types.SeverityError
	} else if status >= 400 {
		severity = types.SeverityWarning
	}

	return &types.Event{
		Timestamp:  ts,
		ServerID:   p.serverID,
		EventType:  types.EventNginxRequest,
		Severity:   severity,
		SourceIP:   m[1],
		Username:   m[2],
		Message:    m[4] + " " + m[5] + " -> " + m[6],
		RawLog:     line,
		Metadata:   map[string]interface{}{
			"method":   m[4],
			"uri":      m[5],
			"status":   status,
			"bytes":    m[7],
			"referer":  m[8],
			"useragent": m[9],
		},
	}
}

var nginxErrorPattern = regexp.MustCompile(`^\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+\[(\w+)\]\s+\d+#\d+:\s+(.*)`)

func (p *NginxParser) ParseError(line string, ts time.Time) *types.Event {
	m := nginxErrorPattern.FindStringSubmatch(line)
	if len(m) < 3 {
		return nil
	}

	severity := types.SeverityWarning
	switch m[1] {
	case "error":
		severity = types.SeverityError
	case "crit", "alert", "emerg":
		severity = types.SeverityCritical
	}

	return &types.Event{
		Timestamp: ts,
		ServerID:  p.serverID,
		EventType: types.EventNginxError,
		Severity:  severity,
		Message:   m[2],
		RawLog:    line,
	}
}

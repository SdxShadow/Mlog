package application

import (
	"regexp"
	"strconv"
	"time"

	"github.com/SdxShadow/Mlog/pkg/types"
)

type ApacheParser struct {
	serverID string
}

func NewApacheParser(serverID string) *ApacheParser {
	return &ApacheParser{serverID: serverID}
}

var apacheAccessPattern = regexp.MustCompile(`^(\S+)\s+\S+\s+\S+\s+\[([^\]]+)\]\s+"(\S+)\s+(\S+)\s+\S+"\s+(\d+)\s+(\d+|-)`)

func (p *ApacheParser) ParseAccess(line string, ts time.Time) *types.Event {
	m := apacheAccessPattern.FindStringSubmatch(line)
	if len(m) < 7 {
		return nil
	}

	status, _ := strconv.Atoi(m[5])
	severity := types.SeverityInfo
	if status >= 500 {
		severity = types.SeverityError
	} else if status >= 400 {
		severity = types.SeverityWarning
	}

	return &types.Event{
		Timestamp:  ts,
		ServerID:   p.serverID,
		EventType:  types.EventApacheRequest,
		Severity:   severity,
		SourceIP:   m[1],
		Message:    m[3] + " " + m[4] + " -> " + m[5],
		RawLog:     line,
		Metadata:   map[string]interface{}{
			"method": m[3],
			"uri":    m[4],
			"status": status,
			"bytes":  m[6],
		},
	}
}

var apacheErrorPattern = regexp.MustCompile(`^\[([A-Z][a-z]{2})\s+([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2})\.\d+\s+(\S+)\s+(\S+)\]\s+\[(\w+)\]\s+(.*)`)

func (p *ApacheParser) ParseError(line string, ts time.Time) *types.Event {
	m := apacheErrorPattern.FindStringSubmatch(line)
	if len(m) < 7 {
		return nil
	}

	severity := types.SeverityWarning
	switch m[5] {
	case "error":
		severity = types.SeverityError
	case "crit", "alert", "emerg":
		severity = types.SeverityCritical
	}

	return &types.Event{
		Timestamp: ts,
		ServerID:  p.serverID,
		EventType: types.EventApacheError,
		Severity:  severity,
		Message:   m[6],
		RawLog:    line,
		Metadata: map[string]interface{}{
			"level": m[5],
		},
	}
}

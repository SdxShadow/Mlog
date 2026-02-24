package ssh

import (
	"regexp"
	"time"

	"github.com/SdxShadow/Mlog/pkg/types"
)

type Parser struct {
	serverID string
}

func New(serverID string) *Parser {
	return &Parser{serverID: serverID}
}

type pattern struct {
	regex   *regexp.Regexp
	handler func([]string, string) *types.Event
}

var patterns = []pattern{
	{
		regexp.MustCompile(`Accepted (password|publickey) for (\S+) from (\S+) port (\d+)`),
		func(m []string, raw string) *types.Event {
			return &types.Event{
				EventType:   types.EventSSHConnected,
				Severity:    types.SeverityInfo,
				Username:    m[2],
				SourceIP:    m[3],
				SourcePort:  toInt(m[4]),
				Message:     "SSH login successful",
				RawLog:      raw,
			}
		},
	},
	{
		regexp.MustCompile(`Failed (password|keyboard-interactive) for (?:invalid user )?(\S+) from (\S+) port (\d+)`),
		func(m []string, raw string) *types.Event {
			return &types.Event{
				EventType:   types.EventSSHFailedAuth,
				Severity:    types.SeverityWarning,
				Username:    m[2],
				SourceIP:    m[3],
				SourcePort:  toInt(m[4]),
				Message:     "SSH login failed",
				RawLog:      raw,
			}
		},
	},
	{
		regexp.MustCompile(`Invalid user (\S+) from (\S+) port (\d+)`),
		func(m []string, raw string) *types.Event {
			return &types.Event{
				EventType:   types.EventSSHFailedAuth,
				Severity:    types.SeverityWarning,
				Username:    m[1],
				SourceIP:    m[2],
				SourcePort:  toInt(m[3]),
				Message:     "SSH invalid user",
				RawLog:      raw,
			}
		},
	},
	{
		regexp.MustCompile(`Disconnected from user (\S+) \[preauth\]`),
		func(m []string, raw string) *types.Event {
			return &types.Event{
				EventType:   types.EventSSHDisconnected,
				Severity:    types.SeverityInfo,
				Username:    m[1],
				Message:     "SSH disconnected",
				RawLog:      raw,
			}
		},
	},
	{
		regexp.MustCompile(`Received disconnect from (\S+) port (\d+)`),
		func(m []string, raw string) *types.Event {
			return &types.Event{
				EventType:   types.EventSSHDisconnected,
				Severity:    types.SeverityInfo,
				SourceIP:    m[1],
				SourcePort:  toInt(m[2]),
				Message:     "SSH disconnected",
				RawLog:      raw,
			}
		},
	},
	{
		regexp.MustCompile(`Connection closed by (\S+) port \d+`),
		func(m []string, raw string) *types.Event {
			return &types.Event{
				EventType:   types.EventSSHDisconnected,
				Severity:    types.SeverityInfo,
				SourceIP:    m[1],
				Message:     "SSH connection closed",
				RawLog:      raw,
			}
		},
	},
}

func (p *Parser) Parse(line string, ts time.Time) *types.Event {
	for _, pat := range patterns {
		m := pat.regex.FindStringSubmatch(line)
		if len(m) > 0 {
			event := pat.handler(m[1:], line)
			event.Timestamp = ts
			event.ServerID = p.serverID
			return event
		}
	}
	return nil
}

func toInt(s string) int {
	var n int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}

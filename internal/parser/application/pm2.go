package application

import (
	"regexp"
	"strings"
	"time"

	"github.com/SdxShadow/Mlog/pkg/types"
)

type PM2Parser struct {
	serverID string
}

func NewPM2Parser(serverID string) *PM2Parser {
	return &PM2Parser{serverID: serverID}
}

var (
	pm2StartPattern   = regexp.MustCompile(`\[\S+\]\s+(?:App name|PM2)\s+(\S+)\s+(?:has been|being) (started|launched)`)
	pm2StopPattern    = regexp.MustCompile(`\[\S+\]\s+(?:App name|PM2)\s+(\S+)\s+has been (stopped|deleted)`)
	pm2RestartPattern = regexp.MustCompile(`\[\S+\]\s+(?:App name|PM2)\s+(\S+)\s+(?:has been restarted|restarting)`)
	pm2ExitPattern    = regexp.MustCompile(`\[\S+\]\s+(?:App name|PM2)\s+(\S+)\s+(?:exited with code|has exited)`)
	pm2ErrorPattern   = regexp.MustCompile(`(Error:|Exception:|ERR_|TypeError:|SyntaxError:)`)
	pm2CrashPattern   = regexp.MustCompile(`(SIGSEGV|SIGABRT|SIGBUS|segmentation fault|heap out of memory)`)
)

func (p *PM2Parser) Parse(line string, ts time.Time) *types.Event {
	lineLower := strings.ToLower(line)

	if pm2StartPattern.MatchString(lineLower) {
		m := pm2StartPattern.FindStringSubmatch(lineLower)
		return &types.Event{
			Timestamp: ts,
			ServerID:  p.serverID,
			EventType: types.EventPM2Start,
			Severity:  types.SeverityInfo,
			Message:   "PM2 process started: " + m[1],
			RawLog:    line,
		}
	}

	if pm2StopPattern.MatchString(lineLower) {
		m := pm2StopPattern.FindStringSubmatch(lineLower)
		return &types.Event{
			Timestamp: ts,
			ServerID:  p.serverID,
			EventType: types.EventPM2Stop,
			Severity:  types.SeverityInfo,
			Message:   "PM2 process stopped: " + m[1],
			RawLog:    line,
		}
	}

	if pm2RestartPattern.MatchString(lineLower) {
		m := pm2RestartPattern.FindStringSubmatch(lineLower)
		return &types.Event{
			Timestamp: ts,
			ServerID:  p.serverID,
			EventType: types.EventPM2Restart,
			Severity:  types.SeverityInfo,
			Message:   "PM2 process restarted: " + m[1],
			RawLog:    line,
		}
	}

	if pm2ExitPattern.MatchString(lineLower) {
		m := pm2ExitPattern.FindStringSubmatch(lineLower)
		return &types.Event{
			Timestamp: ts,
			ServerID:  p.serverID,
			EventType: types.EventPM2Exit,
			Severity:  types.SeverityWarning,
			Message:   "PM2 process exited",
			RawLog:    line,
			Metadata: map[string]interface{}{
				"reason": m[1],
			},
		}
	}

	if pm2CrashPattern.MatchString(lineLower) {
		return &types.Event{
			Timestamp: ts,
			ServerID:  p.serverID,
			EventType: types.EventPM2Crash,
			Severity:  types.SeverityCritical,
			Message:   "PM2 process crashed",
			RawLog:    line,
		}
	}

	if pm2ErrorPattern.MatchString(line) {
		return &types.Event{
			Timestamp: ts,
			ServerID:  p.serverID,
			EventType: types.EventPM2Error,
			Severity:  types.SeverityError,
			Message:   "PM2 error",
			RawLog:    line,
		}
	}

	return nil
}

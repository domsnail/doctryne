package alerts

import (
	"log/slog"
	"time"
	"uuid"
)

type Alert struct {
	Fingerprint uuid.UUID `json:"fingerprint"`
	Environment string    `json:"environment"`
	Service     *string   `json:"service,omitempty"`
	Tool        string    `json:"tool"`
	Host        string    `json:"host"`

	ClientUUID any `json:"client_uuid,omitempty"`
	MessageID  any `json:"message_id,omitempty"`

	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Severity    Severity `json:"severity"`

	StartedAt *time.Time `json:"started_at,omitempty"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
	SeverityUnknown  Severity = "unknown"
)

var slogSeverityMap = map[slog.Level]Severity{
	slog.LevelDebug:    SeverityUnknown,
	slog.LevelInfo:     SeverityInfo,
	slog.LevelInfo + 1: SeverityLow,
	slog.LevelInfo + 2: SeverityLow,
	slog.LevelInfo + 3: SeverityLow,
	slog.LevelWarn:     SeverityMedium,
	slog.LevelWarn + 1: SeverityHigh,
	slog.LevelWarn + 2: SeverityHigh,
	slog.LevelWarn + 3: SeverityHigh,
	slog.LevelError:    SeverityCritical,
}

func NewAlert() Alert {
	return Alert{}
}

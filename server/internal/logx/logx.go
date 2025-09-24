package logx

import (
	"encoding/json"
	"os"
	"time"
)

func Info(event string, fields map[string]any) {
	logWith("info", event, fields)
}
func Error(event string, fields map[string]any) {
	logWith("error", event, fields)
}

func logWith(level, event string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["ts"] = time.Now().UTC().Format(time.RFC3339Nano)
	fields["level"] = level
	fields["event"] = event
	_ = json.NewEncoder(os.Stdout).Encode(fields)
}

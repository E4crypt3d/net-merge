package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// A running tray or `run` may be elevated and cannot be signalled from a normal shell, so a limit
// change travels through a file in ctl/, like the on/off markers. The instance applies it to its
// scheduler and saves limits.json itself.

// LimitRequest is one adapter's new cap, in Mbit/s; 0 removes the cap.
type LimitRequest struct {
	Adapter string  `json:"adapter"`
	Mbps    float64 `json:"mbps"`
}

func limitRequestPath(dir string) string { return filepath.Join(dir, "ctl", "limit") }

// SendLimit hands a cap to a running instance.
func SendLimit(dir string, req LimitRequest) error {
	if err := os.MkdirAll(filepath.Join(dir, "ctl"), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	tmp := limitRequestPath(dir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o666); err != nil {
		return err
	}
	return os.Rename(tmp, limitRequestPath(dir))
}

// TakeLimit consumes a pending request. The file is renamed aside before it is read, so a request
// that arrives at the same moment cannot be deleted unread.
func TakeLimit(dir string) (LimitRequest, bool) {
	p := limitRequestPath(dir)
	taken := p + ".taken"
	os.Remove(taken) // left over if an earlier read failed
	if err := os.Rename(p, taken); err != nil {
		return LimitRequest{}, false
	}
	b, err := os.ReadFile(taken)
	os.Remove(taken)
	if err != nil {
		return LimitRequest{}, false
	}
	var req LimitRequest
	if json.Unmarshal(b, &req) != nil || req.Adapter == "" {
		return LimitRequest{}, false
	}
	return req, true
}

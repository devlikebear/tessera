package lifeline

import "time"

type Heartbeat struct {
	WorkerID string
	TaskID   string
	At       time.Time
}

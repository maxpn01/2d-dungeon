package protocol

import "time"

type ChatEntry struct {
	At   time.Time
	Name string
	Text string
	Mine bool
}

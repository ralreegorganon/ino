package ino

import (
	"time"
)

type Feed struct {
	FeedID        int64     `json:"feedId" db:"feed_id"`
	RemoteAddress string    `json:"remoteAddress" db:"remote_address"`
	Active        bool      `json:"active" db:"active"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

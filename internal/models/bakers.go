package models

import "time"

type Baker struct {
	Address                  string `gorm:"primaryKey;size:50" json:"address"`
	FirstSeen                time.Time `gorm:"not null" json:"first_seen"`
	LastSeen                 time.Time `gorm:"not null" json:"last_seen"`
	TotalDelegationsReceived int64 `gorm:"default:0" json:"total_delegations_received"`
	UniqueDelegators         int `gorm:"default:0" json:"unique_delegators"`
}

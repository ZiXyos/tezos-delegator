package models

import (
	"time"

	"github.com/google/uuid"
)

type Delegation struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Delegator         string    `gorm:"size:50;not null;index:idx_delegations_delegator" json:"delegator"`
	BakerID           string    `gorm:"size:50;not null;index:idx_delegations_baker" json:"baker_id"`
	Amount            int64     `gorm:"not null;index:idx_delegations_amount,sort:desc" json:"amount"`
	Timestamp         time.Time `gorm:"not null;index:idx_delegations_timestamp,sort:desc;index:idx_delegations_date,expression:DATE(timestamp)" json:"timestamp"`
	Level             int64     `gorm:"not null;index:idx_delegations_level" json:"level"`
	OperationHash     *string   `gorm:"size:100;unique" json:"operation_hash"`
	IsNewDelegation   bool      `gorm:"default:false" json:"is_new_delegation"`
	PreviousBaker     *string   `gorm:"size:50" json:"previous_baker"`
	CreatedAt         time.Time `gorm:"default:now()" json:"created_at"`
	IndexedAt         time.Time `gorm:"default:now()" json:"indexed_at"`
	
	Baker Baker `gorm:"foreignKey:BakerID;references:Address" json:"baker,omitempty"`
}

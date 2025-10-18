package domain

import "time"

type DelegationsResponseType struct {
	Timestamp time.Time `json:"timestamp"`
	Amount    int64     `json:"amount"`
	Delegator string    `json:"delegator"`
	Level     int64     `json:"level"`
}
type ApiResponse[T any] struct {
	Data []T `json:"data"`
}

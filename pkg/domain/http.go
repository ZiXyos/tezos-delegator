package domain

import "time"

type Delegations struct {
	timestamp time.Time
	amount    int64
	delegator string
	level     string
}
type Response struct {
	data []Delegations
}

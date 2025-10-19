package domain

type DelegationService interface {
	GetDelegations() ([]TzktApiDelegationsResponse, error)
	GetDelegationsFromLevel(lastLevel int64, limit int) ([]TzktApiDelegationsResponse, error)
}

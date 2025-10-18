package domain

type DelegationService interface {
	GetDelegations() ([]TzktApiDelegationsResponse, error)
}

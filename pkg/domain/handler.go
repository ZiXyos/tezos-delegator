package domain

import "context"

type Handler interface {
	Run(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

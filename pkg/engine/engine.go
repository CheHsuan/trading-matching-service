package engine

import "context"

type Engine interface {
	// Run starts running the engine.
	Run(ctx context.Context) error
}

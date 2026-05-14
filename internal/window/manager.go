package window

import "context"

type Manager struct {
	ctx context.Context
}

func NewManager(ctx context.Context) *Manager {
	return &Manager{ctx: ctx}
}

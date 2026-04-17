package pool

import (
	"github.com/stsgym/vimic2/internal/types"
)

// PoolManagerAdapter wraps PoolManager to implement types.PoolManagerInterface.
// The only mismatch is GetPool which returns *Pool instead of *types.PoolState.
type PoolManagerAdapter struct {
	*PoolManager
}

// NewPoolManagerAdapter creates an adapter
func NewPoolManagerAdapter(pm *PoolManager) *PoolManagerAdapter {
	return &PoolManagerAdapter{PoolManager: pm}
}

// GetPool returns a PoolState instead of Pool
func (a *PoolManagerAdapter) GetPool(name string) (*types.PoolState, error) {
	p, err := a.PoolManager.GetPool(name)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	return &types.PoolState{
		Name:         p.Name,
		TemplatePath: p.TemplateID,
		Capacity:     p.MaxSize,
		Available:    p.MaxSize - p.CurrentSize,
		Busy:         p.CurrentSize,
	}, nil
}

// Verify interface satisfaction
var _ types.PoolManagerInterface = (*PoolManagerAdapter)(nil)
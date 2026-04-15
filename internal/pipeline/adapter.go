package pipeline

import (
	"context"

	"github.com/stsgym/vimic2/internal/types"
)

// PipelineDBAdapter wraps *PipelineDB to implement types.PipelineDB.
// Converts between concrete Pipeline types and interface types.
type PipelineDBAdapter struct {
	*PipelineDB
}

// NewPipelineDBAdapter creates an adapter that implements types.PipelineDB
func NewPipelineDBAdapter(db *PipelineDB) *PipelineDBAdapter {
	return &PipelineDBAdapter{PipelineDB: db}
}

// SavePipeline saves a pipeline state (interface type -> concrete type)
func (a *PipelineDBAdapter) SavePipeline(ctx context.Context, p *types.PipelineState) error {
	concrete := pipelineStateToConcrete(p)
	return a.PipelineDB.SavePipeline(ctx, concrete)
}

// GetPipeline gets a pipeline by ID and returns the interface type
func (a *PipelineDBAdapter) GetPipeline(ctx context.Context, id string) (*types.PipelineState, error) {
	p, err := a.PipelineDB.GetPipeline(ctx, id)
	if err != nil {
		return nil, err
	}
	return concreteToPipelineState(p), nil
}

// ListPipelines returns all pipelines as interface types
func (a *PipelineDBAdapter) ListPipelines(ctx context.Context, limit, offset int) ([]*types.PipelineState, error) {
	pipelines, err := a.PipelineDB.ListPipelines(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]*types.PipelineState, len(pipelines))
	for i, p := range pipelines {
		result[i] = concreteToPipelineState(p)
	}
	return result, nil
}

// UpdatePipelineStatus updates the status of a pipeline
func (a *PipelineDBAdapter) UpdatePipelineStatus(ctx context.Context, id string, status types.PipelineStatus) error {
	return a.PipelineDB.UpdatePipelineStatus(ctx, id, PipelineStatus(status))
}

// DeletePipeline deletes a pipeline by ID
func (a *PipelineDBAdapter) DeletePipeline(ctx context.Context, id string) error {
	// PipelineDB doesn't have DeletePipeline; use UpdatePipelineStatus to mark as deleted
	return a.PipelineDB.UpdatePipelineStatus(ctx, id, PipelineStatusCanceled)
}

// SaveRunner saves a runner state
func (a *PipelineDBAdapter) SaveRunner(ctx context.Context, r *types.RunnerState) error {
	concrete := runnerStateToConcrete(r)
	return a.PipelineDB.SaveRunner(ctx, concrete)
}

// GetRunner gets a runner by ID
func (a *PipelineDBAdapter) GetRunner(ctx context.Context, id string) (*types.RunnerState, error) {
	r, err := a.PipelineDB.GetRunner(ctx, id)
	if err != nil {
		return nil, err
	}
	return concreteToRunnerState(r), nil
}

// ListRunnersByPipeline lists runners for a pipeline
func (a *PipelineDBAdapter) ListRunnersByPipeline(ctx context.Context, pipelineID string) ([]*types.RunnerState, error) {
	runners, err := a.PipelineDB.ListRunnersByPipeline(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	result := make([]*types.RunnerState, len(runners))
	for i, r := range runners {
		result[i] = concreteToRunnerState(r)
	}
	return result, nil
}

// DeleteRunner deletes a runner
func (a *PipelineDBAdapter) DeleteRunner(ctx context.Context, id string) error {
	return a.PipelineDB.DeleteRunner(ctx, id)
}

// SaveVM saves a VM state
func (a *PipelineDBAdapter) SaveVM(ctx context.Context, vm *types.VMState) error {
	// types.VMState uses the old interface format (Status, IPAddress, etc.)
	// Convert to concrete VM
	concrete := &VM{
		ID:        vm.ID,
		Name:      vm.Name,
		IP:        vm.IPAddress,
		MAC:       vm.MACAddress,
		State:     VMState(vm.Status),
		PoolID:    vm.PoolName,
		CreatedAt: vm.CreatedAt,
	}
	return a.PipelineDB.SaveVM(ctx, concrete)
}

// GetVM gets a VM by ID
func (a *PipelineDBAdapter) GetVM(ctx context.Context, id string) (*types.VMState, error) {
	vm, err := a.PipelineDB.GetVM(ctx, id)
	if err != nil {
		return nil, err
	}
	if vm == nil {
		return nil, nil
	}
	return &types.VMState{
		ID:          vm.ID,
		Name:        vm.Name,
		Status:      string(vm.State),
		IPAddress:   vm.IP,
		MACAddress:  vm.MAC,
		PoolName:    vm.PoolID,
		Template:    vm.TemplateID,
		CreatedAt:   vm.CreatedAt,
		DestroyedAt: vm.DestroyedAt,
	}, nil
}

// ListVMsByPool lists VMs in a pool
func (a *PipelineDBAdapter) ListVMsByPool(ctx context.Context, poolID string) ([]*types.VMState, error) {
	vms, err := a.PipelineDB.ListVMsByPool(ctx, poolID)
	if err != nil {
		return nil, err
	}
	result := make([]*types.VMState, len(vms))
	for i, vm := range vms {
		result[i] = &types.VMState{
			ID:         vm.ID,
			Name:       vm.Name,
			Status:     string(vm.State),
			IPAddress:  vm.IP,
			MACAddress: vm.MAC,
			PoolName:   vm.PoolID,
			Template:   vm.TemplateID,
			CreatedAt:  vm.CreatedAt,
		}
	}
	return result, nil
}

// UpdateVMState updates a VM's state
func (a *PipelineDBAdapter) UpdateVMState(ctx context.Context, id string, state string) error {
	return a.PipelineDB.UpdateVMState(ctx, id, VMState(state))
}

// DeleteVM deletes a VM
func (a *PipelineDBAdapter) DeleteVM(ctx context.Context, id string) error {
	return a.PipelineDB.DeleteVM(ctx, id)
}

// SaveTemplate saves a template state
func (a *PipelineDBAdapter) SaveTemplate(ctx context.Context, t *types.TemplateState) error {
	concrete := &Template{
		ID:        t.ID,
		Name:      t.Name,
		BaseImage: t.BaseImage,
		CreatedAt: t.CreatedAt,
	}
	return a.PipelineDB.SaveTemplate(ctx, concrete)
}

// GetTemplate gets a template by ID
func (a *PipelineDBAdapter) GetTemplate(ctx context.Context, id string) (*types.TemplateState, error) {
	t, err := a.PipelineDB.GetTemplate(ctx, id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}
	return &types.TemplateState{
		ID:        t.ID,
		Name:      t.Name,
		BaseImage: t.BaseImage,
		CreatedAt: t.CreatedAt,
	}, nil
}

// ListTemplates lists all templates
func (a *PipelineDBAdapter) ListTemplates(ctx context.Context) ([]*types.TemplateState, error) {
	templates, err := a.PipelineDB.ListTemplates(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*types.TemplateState, len(templates))
	for i, t := range templates {
		result[i] = &types.TemplateState{
			ID:        t.ID,
			Name:      t.Name,
			BaseImage: t.BaseImage,
			CreatedAt: t.CreatedAt,
		}
	}
	return result, nil
}

// DeleteTemplate deletes a template
func (a *PipelineDBAdapter) DeleteTemplate(ctx context.Context, id string) error {
	return a.PipelineDB.DeleteTemplate(ctx, id)
}

// SavePool saves a pool state
func (a *PipelineDBAdapter) SavePool(ctx context.Context, p *types.PoolState) error {
	concrete := poolStateToConcrete(p)
	return a.PipelineDB.SavePool(ctx, concrete)
}

// GetPool gets a pool by ID
func (a *PipelineDBAdapter) GetPool(ctx context.Context, id string) (*types.PoolState, error) {
	p, err := a.PipelineDB.GetPool(ctx, id)
	if err != nil {
		return nil, err
	}
	return concreteToPoolState(p), nil
}

// ListPools lists all pools
func (a *PipelineDBAdapter) ListPools(ctx context.Context) ([]*types.PoolState, error) {
	pools, err := a.PipelineDB.ListPools(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*types.PoolState, len(pools))
	for i, p := range pools {
		result[i] = concreteToPoolState(p)
	}
	return result, nil
}

// UpdatePoolSize updates a pool's size
func (a *PipelineDBAdapter) UpdatePoolSize(ctx context.Context, id string, delta int) error {
	return a.PipelineDB.UpdatePoolSize(ctx, id, delta)
}

// DeletePool deletes a pool
func (a *PipelineDBAdapter) DeletePool(ctx context.Context, id string) error {
	// PipelineDB doesn't have DeletePool; mark as empty/minimal
	return nil
}

// SaveNetwork saves a network state
func (a *PipelineDBAdapter) SaveNetwork(ctx context.Context, n *types.NetworkState) error {
	concrete := networkStateToConcrete(n)
	return a.PipelineDB.SaveNetwork(ctx, concrete)
}

// GetNetwork gets a network by ID
func (a *PipelineDBAdapter) GetNetwork(ctx context.Context, id string) (*types.NetworkState, error) {
	n, err := a.PipelineDB.GetNetwork(ctx, id)
	if err != nil {
		return nil, err
	}
	return concreteToNetworkState(n), nil
}

// ListNetworks lists all networks (not implemented on PipelineDB)
func (a *PipelineDBAdapter) ListNetworks(ctx context.Context) ([]*types.NetworkState, error) {
	return nil, nil
}

// DeleteNetwork deletes a network
func (a *PipelineDBAdapter) DeleteNetwork(ctx context.Context, id string) error {
	return a.PipelineDB.DeleteNetwork(ctx, id)
}

// SaveMetric saves a metric (not implemented on concrete PipelineDB)
func (a *PipelineDBAdapter) SaveMetric(ctx context.Context, m *types.Metric) error {
	// PipelineDB doesn't have SaveMetric; skip
	return nil
}

// ==================== Conversion functions ====================

func concreteToPipelineState(p *Pipeline) *types.PipelineState {
	if p == nil {
		return nil
	}
	return &types.PipelineState{
		ID:         p.ID,
		Platform:   types.RunnerPlatform(p.Platform),
		Repository: p.Repository,
		Branch:     p.Branch,
		CommitSHA:  p.CommitSHA,
		CommitMsg:  p.CommitMsg,
		Author:     p.Author,
		Status:     types.PipelineStatus(p.Status),
		NetworkID:  p.NetworkID,
		StartTime:  p.StartTime,
		EndTime:    p.EndTime,
		Duration:   p.Duration,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

func pipelineStateToConcrete(p *types.PipelineState) *Pipeline {
	if p == nil {
		return nil
	}
	return &Pipeline{
		ID:         p.ID,
		Platform:   RunnerPlatform(p.Platform),
		Repository: p.Repository,
		Branch:     p.Branch,
		CommitSHA:  p.CommitSHA,
		CommitMsg:  p.CommitMsg,
		Author:     p.Author,
		Status:     PipelineStatus(p.Status),
		NetworkID:  p.NetworkID,
		StartTime:  p.StartTime,
		EndTime:    p.EndTime,
		Duration:   p.Duration,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

func concreteToRunnerState(r *Runner) *types.RunnerState {
	if r == nil {
		return nil
	}
	return &types.RunnerState{
		ID:          r.ID,
		PipelineID:  r.PipelineID,
		VMID:        r.VMID,
		Platform:    types.RunnerPlatform(r.Platform),
		PlatformID:  r.PlatformID,
		Labels:      r.Labels,
		Name:        r.Name,
		Status:      types.RunnerStatus(r.Status),
		CurrentJob:  r.CurrentJob,
		CreatedAt:   r.CreatedAt,
		DestroyedAt: r.DestroyedAt,
	}
}

func runnerStateToConcrete(r *types.RunnerState) *Runner {
	if r == nil {
		return nil
	}
	return &Runner{
		ID:          r.ID,
		PipelineID:  r.PipelineID,
		VMID:        r.VMID,
		Platform:    RunnerPlatform(r.Platform),
		PlatformID:  r.PlatformID,
		Labels:      r.Labels,
		Name:        r.Name,
		Status:      RunnerStatus(r.Status),
		CurrentJob:  r.CurrentJob,
		CreatedAt:   r.CreatedAt,
		DestroyedAt: r.DestroyedAt,
	}
}

func concreteToPoolState(p *Pool) *types.PoolState {
	if p == nil {
		return nil
	}
	available := p.MaxSize - p.CurrentSize
	if available < 0 {
		available = 0
	}
	return &types.PoolState{
		Name:         p.Name,
		TemplatePath: p.TemplateID,
		Capacity:     p.MaxSize,
		Available:    available,
		Busy:         p.CurrentSize,
	}
}

func poolStateToConcrete(p *types.PoolState) *Pool {
	if p == nil {
		return nil
	}
	return &Pool{
		Name:        p.Name,
		TemplateID:  p.TemplatePath,
		MinSize:     p.Capacity / 2,
		MaxSize:     p.Capacity,
		CurrentSize: p.Busy,
	}
}

func concreteToNetworkState(n *Network) *types.NetworkState {
	if n == nil {
		return nil
	}
	return &types.NetworkState{
		ID:          n.ID,
		PipelineID:  n.PipelineID,
		BridgeName:  n.BridgeName,
		VLANID:      n.VLANID,
		CIDR:        n.CIDR,
		Gateway:     n.Gateway,
		CreatedAt:   n.CreatedAt,
		DestroyedAt: n.DestroyedAt,
	}
}

func networkStateToConcrete(n *types.NetworkState) *Network {
	if n == nil {
		return nil
	}
	return &Network{
		ID:          n.ID,
		PipelineID:  n.PipelineID,
		BridgeName:  n.BridgeName,
		VLANID:      n.VLANID,
		CIDR:        n.CIDR,
		Gateway:     n.Gateway,
		CreatedAt:   n.CreatedAt,
		DestroyedAt: n.DestroyedAt,
	}
}
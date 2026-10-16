package simulation

func (e eater) protected() bool {
	return e.player != nil && e.player.god || e.ai != nil && (e.ai.Stone || e.ai.Awakening > 0)
}

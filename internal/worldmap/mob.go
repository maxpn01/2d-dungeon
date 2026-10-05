package worldmap

type MobType string

const (
	SpawnGuard     MobType = "spawn-guard"
	DungeonShooter MobType = "shooter"
	DungeonPursuer MobType = "pursuer"
	DungeonHeavy   MobType = "heavy"
	Sentinel       MobType = "sentinel"
)

func (t MobType) Valid() bool {
	switch t {
	case SpawnGuard, DungeonShooter, DungeonPursuer, DungeonHeavy, Sentinel:
		return true
	}
	return false
}

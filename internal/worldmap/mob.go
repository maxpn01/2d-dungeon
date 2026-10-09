package worldmap

type MobType string

const (
	SpawnGuard     MobType = "spawn-guard"
	DungeonShooter MobType = "shooter"
	DungeonPursuer MobType = "pursuer"
	DungeonHeavy   MobType = "heavy"
	GateHeavy      MobType = "gate-heavy"
	Sentinel       MobType = "sentinel"
)

func (t MobType) Valid() bool {
	switch t {
	case SpawnGuard, DungeonShooter, DungeonPursuer, DungeonHeavy, GateHeavy, Sentinel:
		return true
	}
	return false
}

func (t MobType) IsHeavy() bool { return t == DungeonHeavy || t == GateHeavy }

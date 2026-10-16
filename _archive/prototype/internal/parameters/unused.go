package parameters

import (
	"time"
)

const (
	SplitMPCost                = 2.0
	CircleSplitSpeedMultiplier = 1.2
	InitialFruits              = 320
	MaxFruits                  = 500
	FruitSpawnTicks            = TickRate / 10
	SafeZoneMaxMass            = 35.0
	SafeZoneCount              = 12
	SafeZonePlacementAttempts  = 500
	SafeZoneRadiusPadding      = 18
	SafeZoneEdgeMargin         = 80
	SafeZoneMinimumSeparation  = 450
	BossDropLifetimeTicks      = 15 * TickRate
	MassSpeedMinimumMultiplier = 1.0 / 3
	MassSpeedMaximumMultiplier = 4.0 / 3
	MassSpeedExponent          = .18
	SplitMassMovementFraction  = .25
	InitialAIs                 = 5
	InitialAISpawnDelay        = 10
	AIAwakeSeconds             = 120
	AIAsleepSeconds            = 60
	AIFiringDistance           = 180.0
	AISettleTicks              = 2 * TickRate
	AIReactionTicks            = TickRate
	AIAttackIntervalTicks      = 2 * TickRate
	SplitDuration              = 8.0
	SplitLaunchDuration        = .25
	SplitMinimumMass           = 0.5
	SplitMergeDuration         = .35
	SplitTurnResponse          = 7.0
	SplitFollowResponse        = 24.0
	EjectInterval              = .15
	EjectSpeed                 = 650.0
	EjectOwnerDelay            = TickRate * 2
	ContactRatio               = 2.0
	KnockedMass                = 2.0
	BossHitGrace               = 1.2
	BossRecoilDuration         = .26
	BossImpactHold             = .07
	BossRecoilDistance         = 85.0
	BossStartingHP             = 30.0
	BossMaxHP                  = 1000.0
	BossMaxTail                = 100
	BossPartHP                 = 10.0
	BossFirstUnits             = 5
	BossUnitsPerRespawn        = 5
	BossSize                   = 36.0
	BossSpawnMinX              = 700.0
	BossSpawnMaxX              = 1240.0
	BossSpawnMinY              = 60.0
	BossSpawnMaxY              = 645.0
	BossAggroRadius            = 900.0
	BossDashDamage             = 8.0
	BossCountdown              = 3.0
	BossFirstSpawnSeconds      = 300
	BossRespawnIncreaseSeconds = 300
	BossHuntTicks              = 24
	BossWindupTicks            = 12
	BossRushTicks              = 27
	BossRecoverTicks           = 22
	EntityDetectionRadius      = 1200
	EatingContactTime          = 80 * time.Millisecond
	EatingShrinkTime           = 140 * time.Millisecond
)

const EatingSceneTime = EatingContactTime + EatingShrinkTime + 30*time.Millisecond

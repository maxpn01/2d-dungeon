package parameters

import "time"

// World geometry. Changing tile or surface dimensions also changes the
// derived world bounds used by both the client and server.
const (
	Tile           = 40
	MinCellX       = -20
	MaxCellX       = 100
	MinCellY       = -192
	SurfaceCells   = 100
	SurfaceMin     = 10
	SurfaceMax     = SurfaceCells*Tile - 10
	North          = MinCellY * Tile
	ActorSize      = 30
	DescentCircleX = 1000
	DescentCircleY = -5580
	ReturnCircleX  = 1000
	ReturnCircleY  = -6760
)

// Simulation timing and movement.
const (
	TickRate     = 30
	StepSeconds  = 1.0 / TickRate
	PlayerSize   = ActorSize
	SpeedScale   = 100.0
	PlayerSpeed  = 6.0
	MapMin       = float64(SurfaceMin)
	MapMax       = float64(SurfaceMax)
	MinPosition  = MapMin
	MaxPosition  = MapMax - PlayerSize
	InputTimeout = 500 * time.Millisecond
)

// Progression, food, and refuges.
const (
	MaxLevel                   = 100
	StartingHP                 = 5.0
	StartingMaxHP              = 10.0
	FruitXP                    = 0.25
	FruitHP                    = 1.0
	ActorHPFraction            = 0.5
	CircleSplitSpeedMultiplier = 1.2
	FruitSize                  = 15.0
	InitialFruits              = 320
	MaxFruits                  = 500
	FruitSpawnTicks            = TickRate / 10
	SafeZoneMaxMass            = 35.0
	SafeZoneCount              = 12
	SafeZonePlacementAttempts  = 500
	SafeZoneRadiusPadding      = 18
	SafeZoneEdgeMargin         = 80
	SafeZoneMinimumSeparation  = 450
	MassDropLifetimeTicks      = 30 * TickRate
	BossDropLifetimeTicks      = 15 * TickRate
	XPFirstLevelCost           = 5.0
	XPLevelCostIncrease        = 2.0
	HPFirstLevelIncrease       = 4
	MassSpeedMinimumMultiplier = 1.0 / 3
	MassSpeedMaximumMultiplier = 4.0 / 3
	MassSpeedExponent          = .18
	SplitMassMovementFraction  = .25
)

// AI behavior and population.
const (
	InitialAIs            = 5
	InitialAISpawnDelay   = 10
	AIRespawnSeconds      = 45
	AIAwakeSeconds        = 120
	AIAsleepSeconds       = 60
	AISize                = PlayerSize
	AISpeed               = PlayerSpeed
	AIDetectionRadius     = 300.0
	AISettleTicks         = 2 * TickRate
	AIReactionTicks       = TickRate
	AIAttackIntervalTicks = 2 * TickRate
)

// Character abilities and combat.
const (
	SplitDuration       = 8.0
	SplitLaunchDuration = .25
	SplitMinimumMass    = 2.0
	MaxCircleCells      = 16
	SplitMergeDuration  = .35
	SplitTurnResponse   = 7.0
	SplitFollowResponse = 24.0
	BoostMultiplier     = 1.5
	BoostMassPerSecond  = 1.0
	BoostMinimumMass    = 1.0
	EjectInterval       = .25
	EjectOwnerDelay     = TickRate * 2
	DashCost            = 1.0
	DashWindup          = 0.14
	DashBurst           = 0.15
	DashRecovery        = 0.45
	DashSpeedMultiplier = 3.0
	BumpStun            = 0.2
	BumpCost            = 1.0
	BumpKnockback       = 12.0
	ContactRatio        = 2.0
	KnockedMass         = 2.0
	HitProtectionTicks  = 8
	MaxMassDrops        = 128
	BossHitGrace        = 1.2
	BossRecoilDuration  = .26
	BossImpactHold      = .07
	BossRecoilDistance  = 85.0
)

// Boss arena and attack cycle.
const (
	BossStartingHP   = 30.0
	BossMaxHP        = 1000.0
	BossMaxTail      = 100
	BossSize         = 36.0
	BossDashDamage   = 8.0
	BossCountdown    = 3.0
	PortalSeconds    = 0.8
	ArenaLeft        = 12 * Tile
	ArenaRight       = 38 * Tile
	ArenaTop         = -134 * Tile
	ArenaBottom      = -118 * Tile
	BossHuntTicks    = 24
	BossWindupTicks  = 12
	BossRushTicks    = 27
	BossRecoverTicks = 22
)

// Client display, feedback, and network smoothing.
const (
	WindowWidth                   = 1440
	WindowHeight                  = 1000
	RadarWorldSpan        float32 = 3000
	ExplorationRadius             = 400
	EntityDetectionRadius         = 1200
	ChatVisibleLines              = 8
	ChatFadeSeconds               = 10
	ChatFontSize                  = 18
	ChatLineHeight                = 24
	ChatMargin                    = 24
	EatingContactTime             = 80 * time.Millisecond
	EatingShrinkTime              = 140 * time.Millisecond
	EatingSceneTime               = EatingContactTime + EatingShrinkTime + 30*time.Millisecond
	MaxExtrapolation              = 100 * time.Millisecond
	CorrectionSeconds             = 0.05
	TeleportDistance              = 200
	PlayerBufferTicks             = 2.0
	MaxChatEntries                = 50
	ChatCooldown                  = 500 * time.Millisecond
	MessagesPerSecond             = 120.0
)

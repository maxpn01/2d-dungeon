package parameters

import "time"

// World geometry. Changing tile or surface dimensions also changes the
// derived world bounds used by both the client and server.
const (
	Tile         = 40
	MinCellX     = -49
	MaxCellX     = 100
	MinCellY     = -358
	SurfaceCells = 36
	MaxCellY     = SurfaceCells
	SurfaceMin   = 10
	SurfaceMax   = SurfaceCells*Tile - 10
	North        = MinCellY * Tile
	ActorSize    = 30
)

// Simulation timing and movement.
const (
	TickRate     = 30
	StepSeconds  = 1.0 / TickRate
	PlayerSize   = ActorSize
	SpeedScale   = 100.0
	BaseSpeed    = 5.5
	MapMin       = float64(SurfaceMin)
	MapMax       = float64(SurfaceMax)
	MinPosition  = MapMin
	MaxPosition  = MapMax - PlayerSize
	InputTimeout = 500 * time.Millisecond
)

// Progression, food, and refuges.
const (
	MaxLevel                   = 1000
	StartingMP                 = 10.0
	MaxMP                      = 10.0
	FruitMP                    = 3.0
	ManaFoodChance             = 30
	ActorMPFraction            = 0.5
	SplitMPCost                = 2.0
	StartingHP                 = 10.0
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
	XPFirstLevelCost           = 15.0
	XPLevelCostGrowth          = 5.0
	KillXPPerLevel             = 5.0
	MPPerLevelIncrease         = 5
	HPPerLevelIncrease         = 5
	MassSpeedMinimumMultiplier = 1.0 / 3
	MassSpeedMaximumMultiplier = 4.0 / 3
	MassSpeedExponent          = .18
	SplitMassMovementFraction  = .25
)

// AI behavior and population.
const (
	InitialAIs              = 5
	InitialAISpawnDelay     = 10
	AIRespawnSeconds        = 60
	AIAwakeSeconds          = 120
	AIAsleepSeconds         = 60
	AISize                  = PlayerSize
	AISpeed                 = BaseSpeed * MobBaseStatMultiplier
	AIDetectionRadius       = 400.0
	AIAimDelayTicks         = TickRate * 2 / 5
	AIAimResetTicks         = TickRate
	AIAlertSeconds          = 10
	AIFrenzyLevel           = 5
	AIFrenzyBoostTicks      = TickRate
	AIFrenzyBoostCycleTicks = 3 * TickRate
	AIChaseLeash            = 900.0
	AIFiringDistance        = 180.0
	AISettleTicks           = 2 * TickRate
	AIReactionTicks         = TickRate
	AIAttackIntervalTicks   = 2 * TickRate
)

// Character abilities and combat.
const (
	SplitDuration       = 8.0
	SplitLaunchDuration = .25
	SplitMinimumMass    = 0.5
	SplitMergeDuration  = .35
	SplitTurnResponse   = 7.0
	SplitFollowResponse = 24.0
	BoostMultiplier     = 1.5 * 1.25 * 0.75
	BoostMPPerSecond    = 1.0
	// MPRecoveryPerSecond        = 0.1
	MPRecoveryPerSecond = 10000
	ShotMPCost          = 0.5
	ShotCooldown        = 1.0 / BaseAttackRate
	// ShotDamage                 = 2.0
	ShotDamage = 1000
	// BaseAttackRate             = 4.0
	BaseAttackRate             = 1000
	AttackRatePerLevelIncrease = 0.25
	// MobBaseStatMultiplier      = 0.7
	MobBaseStatMultiplier = 0.1
	MobAimAccuracy        = 0.95
	ShotSpeed             = 1400.0
	ShotSize              = 10.0
	ShotLifetime          = 1.5
	SecretWallHits        = 5
	MaxProjectiles        = 512
	BoostMinimumMass      = 1.0
	EjectInterval         = .15
	EjectSpeed            = 650.0
	EjectOwnerDelay       = TickRate * 2
	DashCost              = 1.0
	DashWindup            = 0.14
	DashBurst             = 0.15
	DashRecovery          = 0.45
	DashSpeedMultiplier   = 3.0
	BumpStun              = 0.2
	BumpCost              = 1.0
	BumpKnockback         = 12.0
	ContactRatio          = 2.0
	KnockedMass           = 2.0
	HitProtectionTicks    = 8
	MaxMassDrops          = 128
	BossHitGrace          = 1.2
	BossRecoilDuration    = .26
	BossImpactHold        = .07
	BossRecoilDistance    = 85.0
)

const StatueRechargeSeconds = 60

var PerkRankWeights = [6]int{40, 28, 16, 9, 5, 2}

var PerkBonuses = [11][6]float64{
	{.10, .20, .30, .45, .60, .80},
	{.10, .20, .30, .40, .50, .65},
	{2, 4, 6, 8, 10, 15},
	{2, 4, 6, 8, 10, 15},
	{.03, .06, .09, .12, .15, .20},
	{.05, .10, .15, .20, .25, .35},
	{.20, .35, .50, .70, .90, 1.20},
	{.10, .20, .30, .40, .50, .65},
	{.10, .20, .35, .50, .75, 1.00},
	{.10, .20, .30, .45, .65, 1.00},
	{.15, .30, .45, .65, .85, 1.20},
}

var PerkPenalties = [11][6]float64{
	6: {.05, .08, .11, .14, .17, .20},
	7: {.20, .35, .50, .65, .80, 1.00},
}

// Boss arena and attack cycle.
const (
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
	PortalSeconds              = 1.8
	BossHuntTicks              = 24
	BossWindupTicks            = 12
	BossRushTicks              = 27
	BossRecoverTicks           = 22
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
	MatchCountdownSeconds         = 3
	ResultsDelaySeconds           = 3
)

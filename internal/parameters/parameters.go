package parameters

import "time"

// World geometry. Changing tile or surface dimensions also changes the
// derived world bounds used by both the client and server.
const (
	Tile         = 40
	MinCellX     = -49
	MaxCellX     = 112
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
	BaseSpeed    = 5.0
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
	XPLevelCostGrowth          = 10.0
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
	InitialAIs                     = 5
	InitialAISpawnDelay            = 10
	AIRespawnSeconds               = 60
	AIAwakeSeconds                 = 120
	AIAsleepSeconds                = 60
	AISize                         = PlayerSize
	AIDetectionRadius              = 400.0
	AIAimDelayTicks                = TickRate * 2 / 5
	AIAimResetTicks                = TickRate
	AIAlertSeconds                 = 10
	AIFrenzyLevel                  = 5
	AIFrenzyBoostTicks             = TickRate
	AIFrenzyBoostCycleTicks        = 3 * TickRate
	AIChaseLeash                   = 900.0
	AIFiringDistance               = 180.0
	AISettleTicks                  = 2 * TickRate
	AIReactionTicks                = TickRate
	AIAttackIntervalTicks          = 2 * TickRate
	AIDeathsPerLevel               = 2
	AIPerkLevelInterval            = 4
	SentinelLevelOffset            = 15
	SentinelDashBurst              = 0.28
	SentinelAttackPrediction       = 0.75
	SentinelSwordTrackingDuration  = 0.18
	SentinelSwordTurnRate          = 2.5
	SentinelSwordAdvanceMultiplier = 0.8
	SentinelConcurrentAttackers    = 3
	SentinelSpeedMultiplier        = 1.15
	ShooterStrafeTurnCooldownTicks = TickRate / 2
	SentinelFirstPerkLevel         = 14
	ShooterMinimumDistance         = 260.0
	ShooterMaximumDistance         = 350.0
	ShooterLostSightTicks          = TickRate / 2
	ShooterStrafeSpeed             = 0.6
	ShooterStrafeMinimumTicks      = TickRate
	ShooterStrafeVariationTicks    = TickRate
	HeavyStompCooldownTicks        = 3 * TickRate
	HeavyStompReach                = 260.0
	HeavyStompWaveSpeed            = 700.0
	HeavyStompWaveSize             = 42.0
	HeavyStompWaveLifetime         = HeavyStompReach / HeavyStompWaveSpeed
	HeavyStompWindup               = 0.28
	HeavyStompAimLock              = 0.1
	HeavyStompDuration             = 0.65
	HeavyStompStun                 = 0.8
	HeavyStompCost                 = 1.0
	HeavyStompDamageMultiplier     = 1.5
	PursuerSpeedMultiplier         = 1.35
	HeavySizeMultiplier            = 1.2
	GateHeavySizeMultiplier        = 1.65
	GateHeavyHealthMultiplier      = 1.2
	GateHeavyDamageMultiplier      = 1.1
	HeavySpeedMultiplier           = 0.8
	HeavyHealthMultiplier          = 1.6
	PursuerMeleeReach              = 55.0
	PursuerMeleeWindup             = 0.15
	PursuerMeleeDuration           = 0.35
	HeavyMeleeReach                = 80.0
	HeavyMeleeWindup               = 0.4
	HeavyMeleeDuration             = 0.75
	HeavyMeleeDamageMultiplier     = 1.5
	HeavyDodgeChanceMultiplier     = 0.4
	AIShotDodgeChance              = 0.25
	AIShotDodgeLookahead           = 0.2
	AIShotDodgeReactionTicks       = TickRate / 10
	AIShotDodgeDurationTicks       = TickRate / 5
	AIShotDodgeCooldownTicks       = TickRate * 3 / 5
)

// Character abilities and combat.
const (
	SplitDuration              = 8.0
	SplitLaunchDuration        = .25
	SplitMinimumMass           = 0.5
	SplitMergeDuration         = .35
	SplitTurnResponse          = 7.0
	SplitFollowResponse        = 24.0
	BoostMultiplier            = 1.5 * 1.25 * 0.75
	BoostMPPerSecond           = 1.0
	MPRecoveryPerSecond        = 0.1
	ShotMPCost                 = 0.25
	ShotCooldown               = 1.0 / BaseAttackRate
	ShotDamage                 = 2.0
	DamagePerLevelIncrease     = 0.25
	SpeedPerLevelIncrease      = 0.01
	BaseAttackRate             = 4.0
	AttackRatePerLevelIncrease = 0.25
	MobDamageMultiplier        = 0.7
	MobAttackRateMultiplier    = 0.8
	MobMovementSpeedMultiplier = 0.8
	MobAimAccuracy             = 0.9
	ShotSpeed                  = 1400.0
	ShotSize                   = 10.0
	ShotLifetime               = 1.5
	SecretWallHits             = 5
	MaxProjectiles             = 512
	BoostMinimumMass           = 1.0
	EjectInterval              = .15
	EjectSpeed                 = 650.0
	EjectOwnerDelay            = TickRate * 2
	DashCost                   = 1.0
	DashWindup                 = 0.14
	DashBurst                  = 0.15
	DashRecovery               = 0.45
	DashSpeedMultiplier        = 3.0
	BumpStun                   = 0.2
	BumpCost                   = 1.0
	BumpKnockback              = 12.0
	ContactRatio               = 2.0
	KnockedMass                = 2.0
	HitProtectionTicks         = 8
	MaxMassDrops               = 128
	BossHitGrace               = 1.2
	BossRecoilDuration         = .26
	BossImpactHold             = .07
	BossRecoilDistance         = 85.0
)

const StatueRechargeSeconds = 60

var PerkRankWeights = [6]int{40, 28, 16, 9, 5, 2}

var PerkBonuses = [10][6]float64{
	{.10, .20, .30, .45, .60, .80},
	{.10, .20, .30, .40, .50, .65},
	{2, 4, 6, 8, 10, 15},
	{2, 4, 6, 8, 10, 15},
	{.03, .06, .09, .12, .15, .20},
	{.10, .20, .30, .40, .50, .65},
	{.20, .35, .50, .70, .90, 1.20},
	{.10, .20, .35, .50, .75, 1.00},
	{.10, .20, .30, .45, .65, 1.00},
	{.15, .30, .45, .65, .85, 1.20},
}

var PerkPenalties = [10][6]float64{
	5: {.05, .10, .15, .20, .25, .30},
	6: {.05, .08, .11, .14, .17, .20},
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
	WindowWidth                      = 1440
	WindowHeight                     = 1000
	RadarWorldSpan           float32 = 3000
	SecretRuneRevealDistance         = 90.0
	SecretRuneTextDistance           = 65.0
	ExplorationRadius                = 400
	WingWallGlowStrength             = 5
	EntityDetectionRadius            = 1200
	ChatVisibleLines                 = 8
	ChatFadeSeconds                  = 10
	ChatFontSize                     = 18
	ChatLineHeight                   = 24
	ChatMargin                       = 24
	EatingContactTime                = 80 * time.Millisecond
	EatingShrinkTime                 = 140 * time.Millisecond
	EatingSceneTime                  = EatingContactTime + EatingShrinkTime + 30*time.Millisecond
	MaxExtrapolation                 = 100 * time.Millisecond
	CorrectionSeconds                = 0.05
	TeleportDistance                 = 200
	PlayerBufferTicks                = 2.0
	MaxChatEntries                   = 50
	ChatCooldown                     = 500 * time.Millisecond
	MessagesPerSecond                = 120.0
	SnapshotRate                     = 15
	LobbySnapshotInterval            = time.Second
	MatchCountdownSeconds            = 3
	ResultsDelaySeconds              = 3
)

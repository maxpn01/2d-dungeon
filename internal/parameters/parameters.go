package parameters

// World geometry. Changing tile or surface dimensions also changes the
// derived world bounds.
const (
	Tile         = 40
	MinCellX     = -57
	MaxCellX     = 326
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
	TickRate    = 30
	StepSeconds = 1.0 / TickRate
	PlayerSize  = ActorSize
	SpeedScale  = 100.0
	BaseSpeed   = 4.5
	MapMin      = float64(SurfaceMin)
	MapMax      = float64(SurfaceMax)
	MinPosition = MapMin
	MaxPosition = MapMax - PlayerSize
)

// Progression, food, and refuges.
const (
	MaxLevel              = 1000
	StartingMP            = 10.0
	MaxMP                 = 10.0
	FruitMP               = 3.0
	ManaFoodChance        = 30
	ActorMPFraction       = 0.5
	StartingHP            = 10.0
	StartingMaxHP         = 10.0
	FruitXP               = 0.25
	FruitHP               = 1.0
	ActorHPFraction       = 0.5
	FruitSize             = 15.0
	MassDropLifetimeTicks = 30 * TickRate
	XPFirstLevelCost      = 15.0
	XPLevelCostGrowth     = 10.0
	KillXPPerLevel        = 5.0
	MPPerLevelIncrease    = 5
	HPPerLevelIncrease    = 5
)

// AI behavior and population.
const (
	AIRespawnSeconds               = 60
	AISize                         = PlayerSize
	AIDetectionRadius              = 400.0
	AIAimDelayTicks                = TickRate * 2 / 5
	AIAimResetTicks                = TickRate
	AIAlertSeconds                 = 10
	AIFrenzyLevel                  = 5
	AIFrenzyBoostTicks             = TickRate
	AIFrenzyBoostCycleTicks        = 3 * TickRate
	AIChaseLeash                   = 900.0
	AIDeathsPerLevel               = 2
	AIPerkLevelInterval            = 4
	SentinelLevelOffset            = 15
	SentinelDashBurst              = 0.28
	MeleeAttackPrediction          = .9
	SentinelSwordTrackingDuration  = .1
	SentinelSwordTurnRate          = 5.5
	SentinelSwordAdvanceMultiplier = 1.0
	SentinelSpeedMultiplier        = 1.15
	ShooterStrafeTurnCooldownTicks = TickRate / 2
	SentinelFirstPerkLevel         = 14
	ShooterMinimumDistance         = 260.0
	ShooterMaximumDistance         = 350.0
	ShooterLostSightTicks          = TickRate / 2
	ShooterStrafeSpeed             = 0.6
	ShooterStrafeMinimumTicks      = TickRate
	ShooterStrafeVariationTicks    = TickRate
	HeavyStompCooldownTicks        = TickRate * 12 / 5
	GateHeavyStompReach            = 420.0
	GateHeavyStompWaveSize         = 80.0
	GateHeavyStompDamageMultiplier = 2.0
	GateHeavyStompStun             = 1.2
	HeavyStompReach                = 260.0
	HeavyStompWaveSpeed            = 1000.0
	HeavyStompWaveSize             = 56.0
	HeavyStompWaveLifetime         = HeavyStompReach / HeavyStompWaveSpeed
	HeavyStompWindup               = .2
	HeavyStompAimLock              = .05
	HeavyStompDuration             = .48
	HeavyStompStun                 = 0.8
	HeavyStompCost                 = 1.0
	HeavyStompDamageMultiplier     = 1.5
	PursuerSpeedMultiplier         = 2.4
	HeavySizeMultiplier            = 1.2
	GateHeavySizeMultiplier        = 1.65
	GateHeavyHealthMultiplier      = 1.2
	GateHeavyDamageMultiplier      = 1.1
	HeavySpeedMultiplier           = 0.8
	HeavyHealthMultiplier          = 1.6
	PursuerMeleeReach              = 68.0
	PursuerMeleeWindup             = .1
	PursuerMeleeDuration           = .28
	HeavyMeleeReach                = 100.0
	HeavyMeleeWindup               = .22
	HeavyMeleeDuration             = .5
	MeleeAimLock                   = .05
	MeleeTurnRate                  = 6.0
	PursuerMeleeActive             = .1
	HeavyMeleeActive               = .12
	SentinelMeleeActive            = .1
	PursuerMeleeAdvanceMultiplier  = 1.15
	HeavyMeleeAdvanceMultiplier    = 1.6
	SentinelMeleeAdvanceMultiplier = 1.2
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
	BoostMultiplier            = 1.5 * 1.25 * 0.75
	BoostMPPerSecond           = 1.0
	MPRecoveryPerSecond        = 0.1
	ShotMPCost                 = 0.25
	ShotCooldown               = 1.0 / BaseAttackRate
	ShotDamage                 = 2.0
	DamagePerLevelIncrease     = 0.25
	SpeedPerLevelIncrease      = 0.01
	BaseAttackRate             = 5.0
	AttackRatePerLevelIncrease = 0.25
	MobDamageMultiplier        = 0.5
	MobAttackRateMultiplier    = 0.7
	MobShotSpeedMultiplier     = 0.9
	MobMovementSpeedMultiplier = 0.5
	MobAimAccuracy             = 0.9
	ShotSpeed                  = 1800.0
	ShotSize                   = 10.0
	ShotLifetime               = 1.5
	SecretWallHits             = 5
	MaxProjectiles             = 512
	BoostMinimumMass           = 1.0
	DashCost                   = 1.0
	DashWindup                 = 0.14
	DashBurst                  = 0.15
	DashRecovery               = 0.45
	DashSpeedMultiplier        = 3.0
	BumpStun                   = 0.2
	BumpCost                   = 1.0
	BumpKnockback              = 12.0
	HitProtectionTicks         = 8
	MaxMassDrops               = 128
)

const StatueRechargeSeconds = 60

// Forest floor.
const (
	ForestOriginX          = 122
	ForestOriginY          = -230
	ForestPatchRadius      = 3
	ForestPatchCapacity    = 8
	ForestPatchRespawnSecs = 30
	FruitTreeCapacity      = 2
	FruitTreeRespawnSecs   = 45
)

// Forest snakes.
const (
	ForestSnakeCount           = 20
	SnakeRouteStops            = 3
	SnakeLevel                 = 4
	SnakeHealthMultiplier      = 2.5
	SnakeMinStartSegments      = 3
	SnakeMaxStartSegments      = 4
	SnakeSegmentSpacing        = 22.0
	SnakeSegmentSize           = 24.0
	SnakeSpeedMultiplier       = 1.6
	SnakePatrolSpeed           = 0.55
	SnakeDetectionRadius       = 320.0
	SnakeChaseLeash            = 700.0
	SnakeStopPauseTicks        = 2 * TickRate
	SnakeLegTimeoutTicks       = 40 * TickRate
	SnakeBiteReach             = 58.0
	SnakeBiteWindup            = .3
	SnakeBiteDuration          = .65
	SnakeBiteActive            = .1
	SnakeBiteAdvance           = 2.2
	SnakeBiteDamage            = 1.4
	SnakeBitePullback          = 10.0
	SnakeRetreatMinTicks       = TickRate
	SnakeRetreatVariationTicks = TickRate * 3 / 4
	SnakeRetreatDistance       = 170.0
	SnakeRetreatSpeed          = .75
	SnakeRetreatProbe          = 70.0
	SnakeFlinchDamage          = .2
	SnakeFlinchDecay           = .25
	SnakeWeaveSeconds          = 1.1
	SnakeWeaveStrength         = .6
	SnakePoisonSeconds         = 4.0
	SnakeMaxSegments           = 20
	SnakeFoodXP                = 20.0
	SnakeForageRadius          = 170.0
	SnakeForageTimeoutTicks    = 6 * TickRate
	SnakeSpecialGapTicks       = TickRate * 3 / 2
	SnakeSpitMinRange          = 130.0
	SnakeSpitRange             = 340.0
	SnakeSpitWindup            = .6
	SnakeSpitAimLock           = .15
	SnakeSpitSpeed             = 520.0
	SnakeSpitSize              = 18.0
	SnakeSpitDamage            = .7
	SnakeSpitPoisonSeconds     = 3.0
	SnakeSpitCost              = 1.0
	SnakeSpitCooldownTicks     = TickRate * 4
	SnakeSpitPullback          = 14.0
	SnakeSlitherMinRange       = 140.0
	SnakeSlitherRange          = 300.0
	SnakeSlitherCoil           = .65
	SnakeSlitherAimLock        = .2
	SnakeSlitherSurge          = .5
	SnakeSlitherSpeed          = 2.8
	SnakeSlitherTurnRate       = 1.4
	SnakeSlitherRecovery       = .45
	SnakeSlitherDamage         = 1.2
	SnakeSlitherCooldownTicks  = TickRate * 6
)

// Status effects.
const (
	PoisonTickSeconds  = .5
	PoisonDamageRatio  = .15
	PoisonSegmentBonus = .04
)

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

const PortalSeconds = 1.8

// Display and feedback.
const (
	WindowWidth                      = 1440
	WindowHeight                     = 1000
	RadarWorldSpan           float32 = 3000
	SecretRuneRevealDistance         = 90.0
	SecretRuneTextDistance           = 65.0
	ExplorationRadius                = 400
	WingExplorationRadius            = 480
	WingOverviewMargin               = 3
	WingWallGlowStrength             = .8
	ResultsDelaySeconds              = 3
	FloorBannerSeconds               = 5
)

const (
	AlertMarkerSeconds    = .9
	ScreenShakeSeconds    = .22
	HurtShakeStrength     = 4.0
	StompShakeStrength    = 6.0
	DeathShakeStrength    = 10.0
	DeathShakeSeconds     = .6
	DeathSequenceSeconds  = 1.3
	PlayerDeathRingRadius = 70.0
	DeathBurstParticles   = 14
	DeathBurstDistance    = 46.0
	DamageNumberLift      = 125.0
)

const (
	DefaultSoundVolume           float32 = .3
	DungeonWelcomeSoundVolume    float32 = .3
	HintSoundVolume              float32 = .08
	MenuClickSoundVolume         float32 = .18
	MenuClickSoundPitch          float32 = .88
	FoodSoundVolume              float32 = .06
	PlayerShotSoundVolume        float32 = .1
	MobShotSoundVolume           float32 = .05
	ShotSoundRadius                      = 700.0
	PlayerShotSoundInterval              = .06
	MobShotSoundInterval                 = .16
	VictorySoundVolume           float32 = .04
	FireSoundVolume              float32 = .08
	HeavyStompSoundVolume        float32 = .12
	MeleeSwingSoundVolume        float32 = .12
	HeavyMeleeSoundVolume        float32 = .14
	SentinelAwakeningSoundVolume float32 = .09
	LibraryDiscoverySoundVolume  float32 = .25
	SealSoundVolume              float32 = .08
	OrbAwakeningSoundVolume      float32 = .1
	FloorAmbienceVolume          float32 = .22
	AmbienceDroneGain                    = 1.0
	AmbienceAirGain                      = .3
	AmbienceBellGain                     = 1.0
	AmbienceWhisperGain                  = .04
	GateSoundVolume              float32 = .2
	EntranceGateSoundVolume      float32 = .16
)

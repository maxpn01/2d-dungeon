package protocol

import (
	"errors"
	"math"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/physics"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"
)

func ValidateSnapshot(snapshot Snapshot, localID string) error {
	if snapshot.BriefingActive && (snapshot.MatchStarted || snapshot.StartCountdown != 0 || len(snapshot.Players) != simulation.MaxPlayers) {
		return errors.New("server sent invalid briefing state")
	}
	if snapshot.FinishedTick > snapshot.Tick || snapshot.WinnerID == "" && snapshot.FinishedTick != 0 {
		return errors.New("server sent invalid round finish time")
	}
	if snapshot.StartCountdown < 0 || snapshot.StartCountdown > parameters.MatchCountdownSeconds || snapshot.MatchStarted && snapshot.StartCountdown != 0 || snapshot.StartCountdown > 0 && len(snapshot.Players) != simulation.MaxPlayers {
		return errors.New("server sent invalid match countdown")
	}
	if snapshot.DungeonLayout != worldmap.Default().LayoutID() || snapshot.DungeonLayout == "" {
		return errors.New("server sent a different dungeon layout")
	}
	if err := validateObjective(snapshot); err != nil {
		return err
	}
	if err := validatePlayers(snapshot.Players, localID); err != nil {
		return err
	}
	if err := validateFruits(snapshot.Fruits); err != nil {
		return err
	}
	if err := validateProjectiles(snapshot.Projectiles); err != nil {
		return err
	}
	if err := validateBoss(snapshot.Boss); err != nil {
		return err
	}
	if err := validateAIs(snapshot.AIs); err != nil {
		return err
	}
	if err := validateStatues(snapshot.Statues); err != nil {
		return err
	}
	if len(snapshot.SafeZones) > simulation.SafeZoneCount {
		return errors.New("server sent too many refuges")
	}
	for _, zone := range snapshot.SafeZones {
		for _, v := range []float64{zone.X, zone.Y, zone.Radius} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return errors.New("server sent invalid refuge")
			}
		}
		if zone.Radius <= 0 || zone.Radius > 100 || zone.X-zone.Radius < simulation.MapMin || zone.Y-zone.Radius < simulation.MapMin || zone.X+zone.Radius > simulation.MapMax || zone.Y+zone.Radius > simulation.MapMax {
			return errors.New("server sent invalid refuge bounds")
		}
	}
	return nil
}

func validatePlayers(players []Player, localID string) error {
	if len(players) > simulation.MaxPlayers {
		return errors.New("server sent too many players")
	}
	ids := make(map[string]bool, len(players))
	for _, player := range players {
		if player.ID == "" || ids[player.ID] {
			return errors.New("server sent invalid player IDs")
		}
		ids[player.ID] = true
		if !validStats(player.Stats, simulation.BaseSpeed) {
			return errors.New("server sent invalid player stats")
		}
		for _, value := range []float64{player.X, player.Y, player.Size} {
			if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > math.MaxFloat32 {
				return errors.New("server sent an invalid player position or size")
			}
		}
		if player.Character != "" && simulation.NormalizeCharacter(player.Character) != player.Character {
			return errors.New("server sent invalid character")
		}
		if player.Stats.HP+player.Split.Mass() > player.Stats.MaxHP()+1e-9 {
			return errors.New("server sent invalid total HP")
		}
		for _, cell := range player.Split {
			if !cell.Active {
				continue
			}
			if player.Character != simulation.Circle || !validStats(cell.Stats, simulation.BaseSpeed) || cell.Stats.HP <= 0 || cell.Stats.Level != player.Stats.Level || cell.Stats.XP != player.Stats.XP {
				return errors.New("server sent invalid split stats")
			}
			for _, v := range []float64{cell.X, cell.Y, cell.Remaining, cell.Launch} {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return errors.New("server sent invalid split cell")
				}
			}
			if cell.Remaining < 0 || cell.Remaining > simulation.SplitDuration || cell.Launch < 0 || cell.Launch > simulation.SplitLaunchDuration || !cell.Direction.Valid() || !worldmap.Default().Walkable(cell.X, cell.Y, simulation.PlayerSize, simulation.PlayerSize) || !worldmap.Default().CanAccessEntrance(player.EntranceID, cell.X, cell.Y, simulation.PlayerSize) {
				return errors.New("server sent invalid split position or timer")
			}
		}
		if err := validateFlight(player); err != nil {
			return err
		}
		if _, ok := worldmap.Default().Gate(player.EntranceID); !ok || player.EntranceGate.ID != player.EntranceID || !worldmap.Default().CanAccessEntrance(player.EntranceID, player.X, player.Y, simulation.PlayerSize) {
			return errors.New("server sent invalid private entrance ownership")
		}
		gate := player.EntranceGate
		if gate.ID != "" {
			if _, ok := worldmap.Default().Gate(gate.ID); !ok {
				return errors.New("server sent unknown entrance gate")
			}
		}
		if math.IsNaN(gate.Open) || math.IsInf(gate.Open, 0) || gate.Open < 0 || gate.Open > 1 || math.IsNaN(gate.Remaining) || math.IsInf(gate.Remaining, 0) || gate.Remaining < 0 || gate.Remaining > 6 || gate.Side < -1 || gate.Side > 1 {
			return errors.New("server sent invalid entrance gate state")
		}
		if math.IsNaN(player.EjectCooldown) || math.IsInf(player.EjectCooldown, 0) || player.EjectCooldown < 0 || player.EjectCooldown > simulation.EjectInterval {
			return errors.New("server sent invalid eject cooldown")
		}
		if math.IsNaN(player.ShotCooldown) || math.IsInf(player.ShotCooldown, 0) || player.ShotCooldown < 0 || player.ShotCooldown > max(parameters.ShotCooldown, 1/player.Stats.AttackRate)+1e-9 {
			return errors.New("server sent invalid shot cooldown")
		}
		if !validCombat(player.Combat) {
			return errors.New("server sent invalid combat state")
		}
		if player.Size <= 0 {
			return errors.New("server sent an invalid player size")
		}
	}
	if !ids[localID] {
		return errors.New("local player is missing from the world")
	}
	return nil
}

func validateFruits(fruits []Fruit) error {
	if len(fruits) > simulation.MaxFruits+simulation.MaxMassDrops {
		return errors.New("server sent too many fruits")
	}
	ids := make(map[string]bool, len(fruits))
	for _, fruit := range fruits {
		if fruit.ID == "" || ids[fruit.ID] {
			return errors.New("server sent invalid fruit IDs")
		}
		ids[fruit.ID] = true
		if math.IsNaN(fruit.MP) || math.IsInf(fruit.MP, 0) || fruit.MP < 0 || fruit.MP > simulation.MaxMP || fruit.MP > 0 && fruit.HP != 0 {
			return errors.New("server sent invalid mana pickup")
		}
		for _, v := range []float64{fruit.X, fruit.Y, fruit.Size} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return errors.New("server sent invalid fruit coordinates")
			}
		}
		maxDropHP := 1.0
		if fruit.Source == "boss" {
			maxDropHP = simulation.BossPartHP
		}
		if fruit.Source != "" && (math.IsNaN(fruit.HP) || fruit.HP < 0 || fruit.HP == 0 && fruit.MP == 0 || fruit.HP > maxDropHP) {
			return errors.New("server sent invalid loot pickup")
		}
		if fruit.Size <= 0 || !worldmap.Default().Walkable(fruit.X, fruit.Y, fruit.Size, fruit.Size) {
			return errors.New("server sent fruit outside the map")
		}
	}
	return nil
}

func validStats(stats PlayerStats, baseSpeed float64) bool {
	for _, rank := range stats.Perks {
		if rank < 0 || rank > simulation.MaxPerkRank {
			return false
		}
	}
	for _, value := range []float64{stats.XP, stats.HP, stats.MP, stats.Speed, stats.Damage, stats.AttackRate, stats.BoostMultiplier, stats.BoostCost} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	if stats.MP < 0 || stats.MP > stats.MaxMP() {
		return false
	}
	if stats.Level < 1 || stats.Level > simulation.MaxLevel || stats.XP < 0 || stats.HP < 0 || stats.HP > stats.MaxHP() || stats.Speed <= 0 || stats.Speed > baseSpeed*2 || stats.Damage <= 0 || stats.Damage > max(100, parameters.ShotDamage*(1+simulation.PerkBonus(simulation.DamagePerk, simulation.MaxPerkRank)+simulation.PerkBonus(simulation.HeavyPerk, simulation.MaxPerkRank))) || stats.AttackRate <= 0 || stats.AttackRate > max(simulation.TickRate, simulation.AttackRateForLevel(stats.Level)*(1+simulation.PerkBonus(simulation.AttackRatePerk, simulation.MaxPerkRank))) || stats.BoostMultiplier < 1 || stats.BoostMultiplier > 10 || stats.BoostCost <= 0 || stats.BoostCost > 10 {
		return false
	}
	if stats.Level == simulation.MaxLevel {
		return stats.XP == 0
	}
	return stats.XP < simulation.XPForNextLevel(stats.Level)
}

func validateAIs(ais []AI) error {
	if len(ais) > simulation.InitialAIs+len(worldmap.Default().Dens()) {
		return errors.New("server sent too many AIs")
	}
	ids := make(map[string]bool, len(ais))
	for _, ai := range ais {
		if ai.ID == "" || ids[ai.ID] {
			return errors.New("server sent invalid AI IDs")
		}
		ids[ai.ID] = true
		if simulation.NormalizeCharacter(ai.Character) != ai.Character && ai.Character != simulation.Triangle {
			return errors.New("server sent an invalid AI character")
		}
		if !validCombat(ai.Combat) {
			return errors.New("server sent invalid AI combat state")
		}
		if !validStats(ai.Stats, simulation.AISpeed) {
			return errors.New("server sent invalid AI stats")
		}
		if ai.Stats.HP+ai.Split.Mass() > ai.Stats.MaxHP()+1e-9 {
			return errors.New("server sent invalid AI total HP")
		}
		for _, cell := range ai.Split {
			if !cell.Active {
				continue
			}
			if ai.Character != simulation.Circle || !validStats(cell.Stats, simulation.AISpeed) || cell.Stats.HP <= 0 || cell.Stats.Level != ai.Stats.Level || cell.Stats.XP != ai.Stats.XP {
				return errors.New("server sent invalid AI split stats")
			}
			for _, v := range []float64{cell.X, cell.Y, cell.Remaining, cell.Launch} {
				if math.IsNaN(v) || math.IsInf(v, 0) {
					return errors.New("server sent invalid AI split cell")
				}
			}
			if cell.Remaining < 0 || cell.Remaining > simulation.SplitDuration || cell.Launch < 0 || cell.Launch > simulation.SplitLaunchDuration || !cell.Direction.Valid() || !worldmap.Default().Walkable(cell.X, cell.Y, simulation.AISize, simulation.AISize) {
				return errors.New("server sent invalid AI split position or timer")
			}
		}
		for _, v := range []float64{ai.X, ai.Y, ai.Size} {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return errors.New("server sent invalid AI coordinates")
			}
		}
		if ai.Size <= 0 || !worldmap.Default().Walkable(ai.X, ai.Y, ai.Size, ai.Size) {
			return errors.New("server sent an AI outside the map")
		}
	}
	return nil
}

func validCombat(c simulation.CombatState) bool {
	if math.IsNaN(c.SwordRemaining) || math.IsInf(c.SwordRemaining, 0) || c.SwordRemaining < 0 || c.SwordRemaining > simulation.SentinelSwordDuration || !c.SwordDirection.Valid() || c.SwordDirection.Dash {
		return false
	}
	for _, v := range []float64{c.RecoilX, c.RecoilY} {
		if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > simulation.BossRecoilDistance+.001 {
			return false
		}
	}
	if math.IsNaN(c.HitGrace) || math.IsInf(c.HitGrace, 0) || c.HitGrace < 0 || c.HitGrace > simulation.BossHitGrace {
		return false
	}
	if !c.Facing.Valid() || !c.Direction.Valid() || c.Facing.Dash || c.Direction.Dash || math.IsNaN(c.Remaining) || math.IsInf(c.Remaining, 0) || c.Remaining < 0 {
		return false
	}
	switch c.Phase {
	case simulation.DashIdle:
		return c.Remaining == 0
	case simulation.DashWinding:
		return c.Remaining <= simulation.DashWindup
	case simulation.DashRushing:
		return c.Remaining <= simulation.DashBurst
	case simulation.DashRecovering:
		return c.Remaining <= simulation.DashRecovery
	case simulation.DashRecoiling:
		return c.Remaining <= simulation.BossRecoilDuration
	case simulation.DashBumped:
		return c.Remaining <= simulation.BumpStun
	default:
		return false
	}
}

func validateBoss(b simulation.Boss) error {
	if b.Phase == "" {
		return nil
	}
	switch b.Phase {
	case "waiting", "gathering", "fighting", "defeated":
	default:
		return errors.New("server sent invalid boss phase")
	}
	switch b.Behavior {
	case "", "hunting", "winding", "rushing", "recovering", "feeding", "patrolling", "fleeing":
	default:
		return errors.New("server sent invalid boss behavior")
	}
	if b.XP < 0 || (b.Level == simulation.MaxLevel && b.XP != 0) || (b.Level < simulation.MaxLevel && b.XP >= simulation.XPForNextLevel(b.Level)) {
		return errors.New("server sent invalid boss XP")
	}
	if b.Chain < 0 || b.Chain > 3 {
		return errors.New("server sent invalid boss combo")
	}
	if len(b.Tail) > simulation.BossMaxTail || b.Level < 1 || b.Level > simulation.MaxLevel || b.HP < 0 || b.HP > b.MaxHP || b.MaxHP > simulation.BossMaxHP || b.MaxHP <= 0 || b.Countdown < 0 || b.Countdown > simulation.BossCountdown || b.SpawnDuration < 0 || b.SpawnTicks < 0 || b.SpawnTicks > b.SpawnDuration || b.Defeats < 0 {
		return errors.New("server sent invalid boss stats")
	}
	for _, v := range []float64{b.X, b.Y, b.HP, b.MaxHP, b.Countdown, b.XP, b.AttackTarget.X, b.AttackTarget.Y} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return errors.New("server sent invalid boss values")
		}
	}
	if b.X < simulation.MapMin || b.Y < simulation.MapMin || b.X+simulation.BossSize > simulation.MapMax || b.Y+simulation.BossSize > simulation.MapMax {
		return errors.New("server sent boss outside surface")
	}
	for _, segment := range b.Tail {
		if math.IsNaN(segment.X) || math.IsNaN(segment.Y) || math.IsInf(segment.X, 0) || math.IsInf(segment.Y, 0) || segment.X < simulation.MapMin || segment.Y < simulation.MapMin || segment.X+simulation.BossSize > simulation.MapMax || segment.Y+simulation.BossSize > simulation.MapMax {
			return errors.New("server sent invalid boss tail")
		}
	}
	return nil
}

func validateProjectiles(shots []simulation.Projectile) error {
	if len(shots) > parameters.MaxProjectiles {
		return errors.New("server sent too many projectiles")
	}
	ids := make(map[string]bool, len(shots))
	for _, shot := range shots {
		if shot.ID == "" || shot.Owner == "" || ids[shot.ID] || !shot.Direction.Valid() {
			return errors.New("server sent invalid projectile")
		}
		ids[shot.ID] = true
		for _, value := range []float64{shot.X, shot.Y, shot.Remaining, shot.Damage, shot.Size} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return errors.New("server sent invalid projectile coordinates")
			}
		}
		if shot.Size < parameters.ShotSize || shot.Size > parameters.ShotSize*(1+simulation.PerkBonus(simulation.ShotSizePerk, simulation.MaxPerkRank)) || shot.Damage <= 0 || shot.Damage > max(100, parameters.ShotDamage*(1+simulation.PerkBonus(simulation.DamagePerk, simulation.MaxPerkRank)+simulation.PerkBonus(simulation.HeavyPerk, simulation.MaxPerkRank))) || shot.Remaining <= 0 || shot.Remaining > parameters.ShotLifetime || !worldmap.Default().Walkable(shot.X, shot.Y, shot.Size, shot.Size) {
			return errors.New("server sent invalid projectile position or lifetime")
		}
	}
	return nil
}

func validateStatues(statues []simulation.StatueState) error {
	points := worldmap.Default().Statues()
	if len(statues) != len(points) {
		return errors.New("server sent invalid statue count")
	}
	seen := map[string]bool{}
	for _, s := range statues {
		if seen[s.ID] || s.Cycle == 0 || len(s.Cards) > 3 || math.IsNaN(s.RechargeRemaining) || math.IsInf(s.RechargeRemaining, 0) || s.RechargeRemaining < 0 || s.RechargeRemaining > parameters.StatueRechargeSeconds || s.RechargeRemaining > 0 && (s.ClaimedBy == "" || worldmap.IsWingStatue(s.ID)) {
			return errors.New("server sent invalid statue")
		}
		seen[s.ID] = true
		found := false
		for _, point := range points {
			at := point.Position()
			if point.ID == s.ID && at.X == s.X && at.Y == s.Y {
				found = true
				break
			}
		}
		if !found {
			return errors.New("server sent invalid statue position")
		}
		families := map[simulation.PerkFamily]bool{}
		for _, card := range s.Cards {
			if !card.Valid() || families[card.Family] {
				return errors.New("server sent invalid perk offer")
			}
			families[card.Family] = true
		}
	}
	return nil
}

func validateObjective(snapshot Snapshot) error {
	if math.IsNaN(snapshot.WingDoorOpen) || math.IsInf(snapshot.WingDoorOpen, 0) || snapshot.WingDoorOpen < 0 || snapshot.WingDoorOpen > 1 {
		return errors.New("server sent invalid secret door animation")
	}
	known := map[string]bool{}
	breakable := map[string]bool{}
	for _, door := range worldmap.Default().SecretDoors() {
		known[door.ID] = true
		breakable[door.ID] = !door.Interactable
	}
	open := map[string]bool{}
	for _, id := range snapshot.OpenSecrets {
		if !known[id] || open[id] {
			return errors.New("server sent invalid secret discovery")
		}
		open[id] = true
	}
	for id, hits := range snapshot.SecretWallHits {
		if !breakable[id] || hits < 1 || hits > parameters.SecretWallHits || hits == parameters.SecretWallHits && !open[id] {
			return errors.New("server sent invalid secret wall damage")
		}
	}
	if !open[worldmap.WingDoor] && snapshot.WingDoorOpen != 0 {
		return errors.New("server sent inconsistent secret door animation")
	}
	if math.IsNaN(snapshot.NorthWingDoorOpen) || math.IsInf(snapshot.NorthWingDoorOpen, 0) || snapshot.NorthWingDoorOpen < 0 || snapshot.NorthWingDoorOpen > 1 || !open[worldmap.NorthID(worldmap.WingDoor)] && snapshot.NorthWingDoorOpen != 0 {
		return errors.New("server sent invalid north door animation")
	}
	position := worldmap.Default().OrbPosition()
	orb := snapshot.Orb
	if orb.X != position.X || orb.Y != position.Y || orb.GuardsRemaining < 0 || orb.GuardsRemaining > 3 || orb.Unlocked != (orb.GuardsRemaining == 0) {
		return errors.New("server sent invalid orb state")
	}
	carrierFound := orb.Carrier == ""
	for _, player := range snapshot.Players {
		if player.WingDoorOpen != snapshot.WingDoorOpen || player.NorthWingDoorOpen != snapshot.NorthWingDoorOpen {
			return errors.New("server sent inconsistent secret door animation")
		}
		if player.HasOrb != (player.ID == orb.Carrier) || player.HasOrb && (player.Dead || !orb.Unlocked) {
			return errors.New("server sent invalid orb carrier")
		}
		if player.ID == orb.Carrier {
			carrierFound = true
		}
		if len(player.OpenSecrets) != len(open) {
			return errors.New("server sent inconsistent secret state")
		}
		seen := map[string]bool{}
		for _, id := range player.OpenSecrets {
			if !open[id] || seen[id] {
				return errors.New("server sent inconsistent secret state")
			}
			seen[id] = true
		}
		if math.IsNaN(player.TeleportCharge) || math.IsInf(player.TeleportCharge, 0) || player.TeleportCharge < 0 || player.TeleportCharge > simulation.PortalSeconds {
			return errors.New("server sent invalid teleport charge")
		}
	}
	if !carrierFound {
		return errors.New("server sent missing orb carrier")
	}
	return nil
}

func validateFlight(p Player) error {
	f := p.Flight
	if math.IsNaN(f.LandingX) || math.IsNaN(f.LandingY) || math.IsInf(f.LandingX, 0) || math.IsInf(f.LandingY, 0) || f.Flying && (!p.HasWings || p.Dead) || f.Returning && !f.Flying {
		return errors.New("server sent invalid flight state")
	}
	if f.Flying {
		geometry := worldmap.Default().WithSecrets(p.OpenSecrets)
		if !geometry.CanFly(physics.Square(p.X, p.Y, simulation.PlayerSize)) || !geometry.Walkable(f.LandingX, f.LandingY, simulation.PlayerSize, simulation.PlayerSize) {
			return errors.New("server sent flight outside the map")
		}
	}
	return nil
}

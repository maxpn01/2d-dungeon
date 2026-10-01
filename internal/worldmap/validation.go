package worldmap

import (
	"fmt"
	"math"
	"raylib-2d-sandbox/internal/parameters"
	"slices"
)

func validateLayout(layout Layout) error {
	if len(layout.Entrances) == 0 {
		return fmt.Errorf("map needs at least one entrance")
	}
	world := Cells(MinCellX, MinCellY, MaxCellX-MinCellX, SurfaceCells-MinCellY)
	roomIDs, passageIDs, feedingIDs, gateIDs, portalIDs := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	statueIDs := map[string]bool{}
	arenaCount := 0
	for _, room := range layout.Rooms {
		if err := uniqueID(roomIDs, room.ID, "room"); err != nil {
			return err
		}
		if !world.ContainsArea(room.Area) {
			return fmt.Errorf("room %s has invalid world bounds", room.ID)
		}
		if room.Floor <= 0 {
			return fmt.Errorf("room %s needs a positive floor number", room.ID)
		}
		local := Cells(0, 0, room.Area.Width, room.Area.Height)
		for _, role := range room.Roles {
			if !slices.Contains([]RoomRole{Entrance, Private, ExplorationRoom, Refuge, BossArena, PortalSanctum}, role) {
				return fmt.Errorf("room %s has unknown role %q", room.ID, role)
			}
		}
		if room.HasRole(BossArena) {
			arenaCount++
		}
		propIDs, poolIDs := map[string]bool{}, map[string]bool{}
		for _, prop := range room.Props {
			if err := uniqueID(propIDs, prop.ID, "room "+room.ID+" prop"); err != nil {
				return err
			}
			if prop.Kind != CoverProp {
				return fmt.Errorf("room %s has unknown prop kind %q", room.ID, prop.Kind)
			}
			if !local.ContainsArea(prop.Area) {
				return fmt.Errorf("prop %s must fit inside room %s", prop.ID, room.ID)
			}
		}
		for _, pool := range room.Pools {
			if err := uniqueID(poolIDs, pool.ID, "room "+room.ID+" pool"); err != nil {
				return err
			}
			if !local.ContainsArea(pool.Area) {
				return fmt.Errorf("pool %s must fit inside room %s", pool.ID, room.ID)
			}
		}
		for _, area := range room.FeedingAreas {
			if err := uniqueID(feedingIDs, area.ID, "feeding area"); err != nil {
				return err
			}
			if area.Radius <= 0 || area.Radius > max(local.Width, local.Height) || !local.Contains(area.At) || !local.ContainsArea(Around(area.At, area.Radius)) {
				return fmt.Errorf("feeding area %s must fit inside room %s", area.ID, room.ID)
			}
			if area.Style != "" && area.Style != TealGarden && area.Style != VioletGarden {
				return fmt.Errorf("feeding area %s has unknown style %q", area.ID, area.Style)
			}
			if area.Food.Capacity < 0 || area.Food.Capacity > 0 && area.Food.Respawn <= 0 {
				return fmt.Errorf("feeding area %s has invalid food settings", area.ID)
			}
			if area.Guard != nil {
				guard := area.Guard
				for _, value := range []float64{guard.Damage, guard.AttackRate} {
					if value < 0 || math.IsNaN(value) || math.IsInf(value, 0) {
						return fmt.Errorf("feeding area %s has invalid combat settings", area.ID)
					}
				}
				if guard.PatrolRadius < 0 || guard.PatrolRadius > max(local.Width, local.Height) || !local.ContainsArea(Around(area.At, guard.PatrolRadius)) || guard.Level <= 0 || guard.Level > parameters.MaxLevel || guard.Respawn <= 0 {
					return fmt.Errorf("feeding area %s has invalid guard settings", area.ID)
				}
			}
		}
		for _, gate := range room.Gates {
			if err := uniqueID(gateIDs, gate.ID, "gate"); err != nil {
				return err
			}
			if gate.At.X < 0 || gate.At.X >= local.Width || gate.At.Y < 0 || gate.At.Y > local.Height || gate.Width <= 0 || gate.Width > local.Width-gate.At.X {
				return fmt.Errorf("gate %s must fit on or inside room %s", gate.ID, room.ID)
			}
		}
		for _, portal := range room.Portals {
			if err := uniqueID(portalIDs, portal.ID, "portal"); err != nil {
				return err
			}
			if !local.Contains(portal.At) {
				return fmt.Errorf("portal %s must fit inside room %s", portal.ID, room.ID)
			}
		}
		for _, statue := range room.Statues {
			if err := uniqueID(statueIDs, statue.ID, "statue"); err != nil {
				return err
			}
			if !local.Contains(statue.At) {
				return fmt.Errorf("statue %s must fit inside room %s", statue.ID, room.ID)
			}
		}
	}
	if arenaCount > 1 {
		return fmt.Errorf("map supports one boss arena")
	}
	for _, room := range layout.Rooms {
		for _, portal := range room.Portals {
			if !portalIDs[portal.Destination] || portal.Destination == portal.ID {
				return fmt.Errorf("portal %s has invalid destination %q", portal.ID, portal.Destination)
			}
		}
	}
	for _, p := range layout.Passages {
		if err := uniqueID(passageIDs, p.ID, "passage"); err != nil {
			return err
		}
		if p.Width <= 0 || p.Width > max(world.Width, world.Height) {
			return fmt.Errorf("passage %s needs a valid width", p.ID)
		}
		area, err := layout.passageArea(p)
		if err != nil {
			return err
		}
		if !world.ContainsArea(area) {
			return fmt.Errorf("passage %s has invalid world bounds", p.ID)
		}
	}
	for _, entry := range layout.Entrances {
		cell, err := layout.resolve(entry)
		if err != nil {
			return fmt.Errorf("entrance: %w", err)
		}
		if !world.Contains(cell) {
			return fmt.Errorf("entrance has invalid world position %v", cell)
		}
	}
	return nil
}

func uniqueID(seen map[string]bool, id, kind string) error {
	if id == "" || seen[id] {
		return fmt.Errorf("%s ID must be nonempty and unique: %q", kind, id)
	}
	seen[id] = true
	return nil
}

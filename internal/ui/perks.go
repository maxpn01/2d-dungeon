package ui

import (
	"fmt"
	"image/color"
	"strings"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/rendering/primitives"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/worldmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type PerkPanel struct {
	id      string
	cycle   uint64
	pending bool
	Closing bool
}

func (p *PerkPanel) Sync(active string, statues []simulation.StatueState) *simulation.StatueState {
	for i := range statues {
		s := &statues[i]
		if s.ID != active || s.ClaimedBy != "" {
			continue
		}
		if p.id != active || p.cycle != s.Cycle {
			*p = PerkPanel{id: active, cycle: s.Cycle}
		}
		return s
	}
	*p = PerkPanel{}
	return nil
}

func perkCardRects(width, height int32, count int) []rl.Rectangle {
	if count == 0 {
		return nil
	}
	gap := float32(18)
	w := min(float32(340), (float32(width)-64-gap*float32(count-1))/float32(count))
	left := (float32(width) - w*float32(count) - gap*float32(count-1)) / 2
	y := float32(height)/2 - 135
	result := make([]rl.Rectangle, count)
	for i := range result {
		result[i] = rl.NewRectangle(left+float32(i)*(w+gap), y, w, 310)
	}
	return result
}

func (p *PerkPanel) Read(s simulation.StatueState, stats simulation.Stats, width, height int32) int {
	if p.pending || p.Closing || !rl.IsWindowFocused() {
		return -1
	}
	if worldmap.IsWingStatue(s.ID) {
		return p.readWingOffer(width, height)
	}
	keys := [...]int32{rl.KeyOne, rl.KeyTwo, rl.KeyThree}
	for i, rect := range perkCardRects(width, height, len(s.Cards)) {
		card := s.Cards[i]
		if card.Rank <= stats.Perks[card.Family] {
			continue
		}
		if rl.IsKeyPressed(keys[i]) || rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), rect) {
			p.pending = true
			return i
		}
	}
	return -1
}

func perkTint(rank int) color.RGBA {
	return [...]color.RGBA{{160, 173, 182, 255}, {101, 186, 133, 255}, {92, 169, 217, 255}, {170, 132, 222, 255}, {236, 170, 85, 255}, {246, 219, 110, 255}}[max(0, min(5, rank-1))]
}

func wrappedText(text string, x, y, width, size int32, tint color.RGBA) int32 {
	line := ""
	for _, word := range strings.Fields(text) {
		candidate := word
		if line != "" {
			candidate = line + " " + word
		}
		if line != "" && primitives.MeasureText(candidate, size) > width {
			primitives.DrawText(line, x, y, size, tint)
			y += size + 8
			line = word
		} else {
			line = candidate
		}
	}
	if line != "" {
		primitives.DrawText(line, x, y, size, tint)
		y += size + 8
	}
	return y
}

func perkPreview(stats simulation.Stats, card simulation.PerkCard, hasWings bool) string {
	next := stats
	if !next.ApplyPerk(card) {
		return "Already owned at this rank or higher"
	}
	switch card.Family {
	case simulation.DamagePerk:
		return fmt.Sprintf("Damage %.2f -> %.2f", stats.Damage, next.Damage)
	case simulation.AttackRatePerk:
		return fmt.Sprintf("Shots/sec %.2f -> %.2f", stats.AttackRate, next.AttackRate)
	case simulation.ManaPerk:
		return fmt.Sprintf("Max MP %.0f -> %.0f. Gain %.0f MP.", stats.MaxMP(), next.MaxMP(), next.MaxMP()-stats.MaxMP())
	case simulation.HealthPerk:
		return fmt.Sprintf("Max HP %.0f -> %.0f. Heal %.0f HP.", stats.MaxHP(), next.MaxHP(), next.MaxHP()-stats.MaxHP())
	case simulation.SpeedPerk:
		return fmt.Sprintf("Base speed %.2f -> %.2f", stats.Speed, next.Speed)
	case simulation.BoostPerk:
		if hasWings {
			return fmt.Sprintf("Flight %.2fx -> %.2fx. MP/sec %.2f -> %.2f.", simulation.FlightSpeedMultiplier(stats), simulation.FlightSpeedMultiplier(next), simulation.FlightMPCost(stats), simulation.FlightMPCost(next))
		}
		return fmt.Sprintf("Shift boost %.2fx -> %.2fx. MP/sec %.2f -> %.2f.", stats.BoostMultiplier, next.BoostMultiplier, stats.BoostCost, next.BoostCost)
	case simulation.HeavyPerk:
		return fmt.Sprintf("Damage %.2f -> %.2f. Speed %.2f -> %.2f.", stats.Damage, next.Damage, stats.Speed, next.Speed)
	case simulation.ManaRecoveryPerk:
		return fmt.Sprintf("MP/sec %.2f -> %.2f", stats.MPRecoveryPerSecond(), next.MPRecoveryPerSecond())
	case simulation.HealthRecoveryPerk:
		return fmt.Sprintf("HP/sec %.2f -> %.2f", stats.HPRecoveryPerSecond(), next.HPRecoveryPerSecond())
	case simulation.ShotSizePerk:
		return fmt.Sprintf("Diameter %.1f -> %.1f. MP/shot %.3f -> %.3f.", stats.ShotSize(), next.ShotSize(), stats.ShotMPCost(), next.ShotMPCost())
	}
	return ""
}

func (p *PerkPanel) Draw(active string, statues []simulation.StatueState, stats simulation.Stats, hasWings bool, width, height int32) {
	s := p.Sync(active, statues)
	if s == nil || p.Closing {
		return
	}
	if worldmap.IsWingStatue(s.ID) {
		p.drawWingOffer(width, height)
		return
	}
	rl.DrawRectangle(0, 0, width, height, color.RGBA{6, 10, 17, 210})
	y := height/2 - 222
	primitives.DrawCenteredText("THE INSCRIPTION", y, 30, primitives.RuneColor)
	primitives.DrawCenteredText(fmt.Sprintf("Choose one blessing. Recharges %d seconds after use.", parameters.StatueRechargeSeconds), y+43, 18, rl.LightGray)
	if len(s.Cards) == 0 {
		primitives.DrawCenteredText("Your blessings are already complete.", height/2, 22, rl.LightGray)
	}
	for i, rect := range perkCardRects(width, height, len(s.Cards)) {
		card := s.Cards[i]
		tint := perkTint(card.Rank)
		enabled := card.Rank > stats.Perks[card.Family]
		hover := enabled && !p.pending && rl.CheckCollisionPointRec(rl.GetMousePosition(), rect)
		background := color.RGBA{23, 31, 43, 255}
		if hover {
			background = color.RGBA{35, 47, 62, 255}
		}
		if !enabled {
			tint = rl.Gray
		}
		rl.DrawRectangleRec(rect, background)
		rl.DrawRectangleLinesEx(rect, 2, tint)
		rl.DrawRectangleRec(rl.NewRectangle(rect.X+2, rect.Y+2, rect.Width-4, 42), rl.Fade(tint, .18))
		x, cy, w := int32(rect.X)+20, int32(rect.Y), int32(rect.Width)-40
		primitives.DrawText(fmt.Sprintf("RANK %s / LV %d", simulation.RankName(card.Rank), card.Rank), x, cy+13, 18, tint)
		primitives.DrawText(simulation.PerkName(card.Family), x, cy+65, 24, rl.RayWhite)
		wrappedText(card.DescriptionFor(hasWings), x, cy+113, w, 20, tint)
		wrappedText(perkPreview(stats, card, hasWings), x, cy+175, w, 17, rl.LightGray)
		owned := stats.Perks[card.Family]
		label := "NEW BLESSING"
		if owned > 0 {
			label = fmt.Sprintf("REPLACES RANK %s", simulation.RankName(owned))
		}
		if !enabled {
			label = "ALREADY OWNED"
			if card.Rank < owned {
				label = "HIGHER RANK OWNED"
			}
		}
		primitives.DrawText(label, x, cy+253, 16, rl.Gray)
		if enabled {
			primitives.DrawText(fmt.Sprintf("[%d] SELECT", i+1), x, cy+279, 18, tint)
		}
	}
	footer := "1-3 or click | Esc to leave | World paused"
	useful := false
	for _, card := range s.Cards {
		if card.Rank > stats.Perks[card.Family] {
			useful = true
			break
		}
	}
	if !useful {
		footer = "No upgrades for your build | Esc to leave this statue available"
	}
	if p.pending {
		footer = "Claiming blessing..."
	}
	primitives.DrawCenteredText(footer, height/2+206, 18, rl.LightGray)
}

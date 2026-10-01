package game

import (
	"fmt"
	"image/color"
	"strings"

	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type perkPanel struct {
	id      string
	cycle   uint64
	pending bool
	closing bool
}

func (p *perkPanel) sync(active string, statues []simulation.StatueState) *simulation.StatueState {
	for i := range statues {
		s := &statues[i]
		if s.ID != active || s.ClaimedBy != "" {
			continue
		}
		if p.id != active || p.cycle != s.Cycle {
			*p = perkPanel{id: active, cycle: s.Cycle}
		}
		return s
	}
	*p = perkPanel{}
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

func (p *perkPanel) read(s simulation.StatueState, stats simulation.Stats, width, height int32) int {
	if p.pending || p.closing || !rl.IsWindowFocused() {
		return -1
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
		if line != "" && measureText(candidate, size) > width {
			drawText(line, x, y, size, tint)
			y += size + 8
			line = word
		} else {
			line = candidate
		}
	}
	if line != "" {
		drawText(line, x, y, size, tint)
		y += size + 8
	}
	return y
}

func perkPreview(stats simulation.Stats, card simulation.PerkCard) string {
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
		return fmt.Sprintf("Speed %.2f -> %.2f", stats.Speed, next.Speed)
	case simulation.BoostPerk:
		return fmt.Sprintf("Shift multiplier %.2f -> %.2f", stats.BoostMultiplier, next.BoostMultiplier)
	case simulation.HeavyPerk:
		return fmt.Sprintf("Damage %.2f -> %.2f. Speed %.2f -> %.2f.", stats.Damage, next.Damage, stats.Speed, next.Speed)
	case simulation.OverdrivePerk:
		return fmt.Sprintf("Boost %.2f -> %.2f. MP/sec %.2f -> %.2f.", stats.BoostMultiplier, next.BoostMultiplier, stats.BoostCost, next.BoostCost)
	}
	return ""
}

func (p *perkPanel) draw(active string, statues []simulation.StatueState, stats simulation.Stats, width, height int32) {
	s := p.sync(active, statues)
	if s == nil || p.closing {
		return
	}
	rl.DrawRectangle(0, 0, width, height, color.RGBA{6, 10, 17, 210})
	y := height/2 - 222
	drawCenteredText("THE INSCRIPTION", y, 30, runeColor)
	drawCenteredText("Choose one blessing. This statue can be claimed by one player.", y+43, 18, rl.LightGray)
	if len(s.Cards) == 0 {
		drawCenteredText("Your blessings are already complete.", height/2, 22, rl.LightGray)
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
		drawText(fmt.Sprintf("RANK %s / LV %d", simulation.RankName(card.Rank), card.Rank), x, cy+13, 18, tint)
		drawText(simulation.PerkName(card.Family), x, cy+65, 24, rl.RayWhite)
		wrappedText(card.Description(), x, cy+113, w, 20, tint)
		wrappedText(perkPreview(stats, card), x, cy+175, w, 17, rl.LightGray)
		owned := stats.Perks[card.Family]
		label := "NEW BLESSING"
		if owned > 0 {
			label = fmt.Sprintf("REPLACES RANK %s", simulation.RankName(owned))
		}
		if !enabled {
			label = "ALREADY OWNED"
		}
		drawText(label, x, cy+253, 16, rl.Gray)
		if enabled {
			drawText(fmt.Sprintf("[%d] SELECT", i+1), x, cy+279, 18, tint)
		}
	}
	footer := "1-3 or click | Esc to leave | The world keeps moving"
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
	drawCenteredText(footer, height/2+206, 18, rl.LightGray)
}

func drawStatues(statues []simulation.StatueState, active string, px, py float64, dead bool, exploration *mapExploration) {
	for _, s := range statues {
		if !exploration.entityVisible(s.X, s.Y, 0) {
			continue
		}
		x, y := int32(s.X), int32(s.Y)
		tint := runeColor
		if s.ClaimedBy != "" {
			tint = color.RGBA{73, 83, 88, 255}
		}
		rl.DrawEllipse(x, y+19, 31, 12, rl.Fade(rl.Black, .5))
		rl.DrawRectangle(x-27, y+10, 54, 12, stone)
		rl.DrawRectangle(x-22, y+6, 44, 7, paleStone)
		rl.DrawRectangle(x-13, y-25, 26, 31, stone)
		rl.DrawRectangle(x-9, y-41, 18, 17, paleStone)
		rl.DrawRectangle(x-23, y-20, 9, 17, paleStone)
		rl.DrawRectangle(x+14, y-20, 9, 17, paleStone)
		rl.DrawRectangle(x-20, y-23, 40, 29, color.RGBA{29, 39, 47, 255})
		rl.DrawRectangleLines(x-20, y-23, 40, 29, paleStone)
		for i := int32(0); i < 3; i++ {
			rl.DrawLine(x-12, y-16+i*6, x+12-i*4, y-16+i*6, tint)
		}
		if s.ClaimedBy == "" {
			rl.DrawCircleLines(x, y-9, 37, rl.Fade(tint, .3))
			if !dead && active == "" && s.Near(px, py) {
				label := "[E] READ INSCRIPTION"
				width := measureText(label, 16)
				rl.DrawRectangle(x-width/2-6, y-71, width+12, 24, rl.Fade(rl.Black, .8))
				drawText(label, x-width/2, y-68, 16, tint)
			}
		} else if !dead && s.Near(px, py) {
			label := "INSCRIPTION CLAIMED"
			drawText(label, x-measureText(label, 14)/2, y-68, 14, rl.Gray)
		}
	}
}

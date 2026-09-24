package game

import (
	"fmt"
	"log"
	"time"

	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
	"raylib-2d-sandbox/internal/terrain"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type multiplayer struct {
	feedback    combatFeedback
	renderedAIs []protocol.AI
	motion      client.Motion
	rendered    []protocol.Player
	session     *client.Session
	view        client.Update
	gameMap     *Map
	errorText   string
	chat        chatBox
	eating      *eatingScene
	otherEating []*eatingScene
}

func newMultiplayer(url, name string, character simulation.Character) *multiplayer {
	return &multiplayer{
		session: client.ConnectWithCharacter(url, name, character),
		gameMap: NewMap(rl.NewVector2(terrain.SurfaceCells*terrain.Tile/10, terrain.SurfaceCells*terrain.Tile/10), rl.NewVector2(10, 10), rl.Black, rl.Gray),
	}
}

func (m *multiplayer) poll() {
	if m.errorText != "" {
		return
	}
	select {
	case update, open := <-m.session.Updates():
		if !open {
			m.errorText = "Connection closed. Please try again."
			return
		}
		if update.Err != nil {
			log.Printf("multiplayer: %v", update.Err)
			if m.view.Ready {
				m.errorText = "Connection lost. The server may be unavailable."
			} else {
				m.errorText = "Could not join the server. Check that it is running."
			}
			return
		}
		m.feedback.multiplayer(time.Now(), update.Players, update.AIs, update.Fruits, update.PlayerID)
		m.startEatingScenes(time.Now(), update)
		m.view = update
		m.gameMap.safeZones = update.SafeZones
		m.motion.Apply(update)
	default:
	}
}

func (m *multiplayer) readInput(dt float32, mapOpen bool, camera rl.Camera2D) {
	if !mapOpen {
		m.chat.update(m)
	}
	direction := protocol.Direction{}
	if m.view.Ready && !m.view.Local.Dead && m.errorText == "" && rl.IsWindowFocused() && !m.chat.open && !mapOpen {
		direction.Eject = rl.IsKeyDown(rl.KeyE)
		direction.Boost = rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
		direction.Dash = rl.IsKeyPressed(rl.KeyLeftShift) || rl.IsKeyPressed(rl.KeyRightShift)
		center := rl.NewVector2(float32(m.view.Local.X+m.view.Local.Size/2), float32(m.view.Local.Y+m.view.Local.Size/2))
		for _, p := range m.rendered {
			if p.ID == m.view.PlayerID {
				center = rl.NewVector2(float32(p.X+p.Size/2), float32(p.Y+p.Size/2))
				break
			}
		}
		aim := mouseAim(camera, center, m.view.Local.Combat.Facing)
		direction.AimX, direction.AimY = float64(aim.X), float64(aim.Y)
		if rl.IsKeyDown(rl.KeyW) {
			direction.Y--
		}
		if rl.IsKeyDown(rl.KeyS) {
			direction.Y++
		}
		if rl.IsKeyDown(rl.KeyA) {
			direction.X--
		}
		if rl.IsKeyDown(rl.KeyD) {
			direction.X++
		}
	}
	m.session.SetInput(direction)
	now := time.Now()
	if predicted, ready := m.session.PredictedPlayer(); ready {
		m.motion.Advance(predicted, float64(dt))
	}
	m.rendered = m.motion.Players(now)
	m.renderedAIs = m.motion.AIs(now)
	if m.view.Ready && !m.view.Local.Dead && m.errorText == "" {
		for _, player := range m.rendered {
			if player.ID == m.view.PlayerID {
				m.gameMap.exploration.reveal(rl.NewVector2(float32(player.X+player.Size/2), float32(player.Y+player.Size/2)))
				break
			}
		}
	}
	active := m.otherEating[:0]
	for _, scene := range m.otherEating {
		if !scene.finished(now) {
			m.rendered, m.renderedAIs = scene.apply(now, m.rendered, m.renderedAIs)
			active = append(active, scene)
		}
	}
	m.otherEating = active
	if m.eating != nil {
		m.rendered, m.renderedAIs = m.eating.apply(now, m.rendered, m.renderedAIs)
	}
}

func (m *multiplayer) follow(camera *rl.Camera2D) {
	if m.eating != nil {
		return
	}
	for _, player := range m.rendered {
		if player.ID == m.view.PlayerID {
			camera.Target = rl.NewVector2(float32(player.X+player.Size/2), float32(player.Y+player.Size/2))
			return
		}
	}
}

func (m *multiplayer) draw(camera rl.Camera2D, window *Window) {
	rl.BeginMode2D(camera)
	m.gameMap.draw(RenderContext{Camera: camera, WindowWidth: window.width, WindowHeight: window.height})
	for _, fruit := range m.view.Fruits {
		if !m.gameMap.exploration.entityVisible(fruit.X, fruit.Y, fruit.Size) {
			continue
		}
		drawFruit(float32(fruit.X), float32(fruit.Y), float32(fruit.Size), massColor(fruit.Source, m.view.PlayerID), fruit.Source != "")
	}
	for _, ai := range m.renderedAIs {
		headVisible := m.gameMap.exploration.entityVisible(ai.X, ai.Y, ai.Size)
		visible := headVisible
		for _, cell := range ai.Split {
			visible = visible || cell.Active && m.gameMap.exploration.entityVisible(cell.X, cell.Y, ai.Size)
		}
		if !visible {
			continue
		}
		mass := ai.Stats.HP + ai.Split.Mass()
		if headVisible {
			drawSafeProtection(ai.X, ai.Y, ai.Size, mass, m.view.SafeZones)
			bounds := drawActor(ai.X, ai.Y, ai.Size, ai.Stats.HP, ai.Combat, rl.Green, ai.Character)
			drawTextAt(fmt.Sprintf("%s Lv.%d %s", ai.ID, ai.Stats.Level, ai.Behavior), rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
		}
		for _, cell := range ai.Split {
			if cell.Active && m.gameMap.exploration.entityVisible(cell.X, cell.Y, ai.Size) {
				drawSafeProtection(cell.X, cell.Y, ai.Size, mass, m.view.SafeZones)
				drawActor(cell.X, cell.Y, ai.Size, cell.Stats.HP, simulation.CombatState{}, rl.Green, ai.Character)
			}
		}
	}
	for _, player := range m.rendered {
		if player.ID != m.view.PlayerID && !m.gameMap.exploration.entityVisible(player.X, player.Y, player.Size) {
			continue
		}
		tint := characterColor(player.Character)
		label := fmt.Sprintf("%s Lv.%d", protocol.NormalizePlayerName(player.Name), player.Stats.Level)
		if player.ID == m.view.PlayerID {
			tint = characterColor(player.Character)
			label += " (you)"
		}
		mass := player.Stats.HP + player.Split.Mass()
		drawSafeProtection(player.X, player.Y, player.Size, mass, m.view.SafeZones)
		bounds := drawActor(player.X, player.Y, player.Size, player.Stats.HP, player.Combat, tint, player.Character)
		for _, cell := range player.Split {
			if cell.Active {
				drawSafeProtection(cell.X, cell.Y, player.Size, mass, m.view.SafeZones)
				drawActor(cell.X, cell.Y, player.Size, cell.Stats.HP, simulation.CombatState{}, tint, player.Character)
			}
		}
		if !player.Dead {
			drawTextAt(label, rl.NewVector2(bounds.X, bounds.Y-22), 18, rl.RayWhite)
		}
	}
	drawEncounter(m.view.Boss, &m.gameMap.exploration)
	for _, p := range m.view.Players {
		drawTeleportCharge(p.X, p.Y, p.TeleportCharge)
	}
	m.feedback.draw(time.Now(), &m.gameMap.exploration)
	rl.EndMode2D()
	drawDungeonLocation(camera.Target.X, camera.Target.Y, window.width, window.height)
	alive := 0
	for _, player := range m.view.Players {
		if !player.Dead {
			alive++
		}
	}
	stats := m.view.Local.Stats
	hp := float32(stats.HP + m.view.Local.Split.Mass())
	if m.view.Local.Dead {
		hp = 0
	}
	drawPlayerHUD(playerHUD{
		name: m.view.Local.Name, avatar: characterColor(m.view.Local.Character), character: m.view.Local.Character,
		level: stats.Level, maxLevel: simulation.MaxLevel,
		hp: hp, maxHP: float32(simulation.MaxHPForLevel(stats.Level)),
		xp: float32(stats.XP), nextXP: float32(simulation.XPForNextLevel(stats.Level)),
		speed: float32(effectivePlayerSpeed(m.view.Local.Character, stats.Speed, stats.HP, m.view.Local.Split)), fruit: len(m.view.Fruits), ais: len(m.view.AIs), players: alive,
	}, window.width)
	drawBossHUD(m.view.Boss, m.view.Local.X, m.view.Local.Y, window.width)
	m.drawMinimap(window.width)
	m.chat.draw(window, m.session.Chat())
}

func (m *multiplayer) startEatingScenes(now time.Time, update client.Update) {
	if update.Local.Dead && m.eating == nil {
		m.eating = newEatingScene(now, update, m.rendered, m.renderedAIs)
	}
	for _, prey := range update.Players {
		if !prey.Dead || prey.ID == update.PlayerID {
			continue
		}
		for _, previous := range m.view.Players {
			if previous.ID == prey.ID && !previous.Dead {
				death := update
				death.Local, death.PlayerID = prey, prey.ID
				m.otherEating = append(m.otherEating, newEatingScene(now, death, m.rendered, m.renderedAIs))
				break
			}
		}
	}
}

package gameplay

import (
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type controls struct{ suppressShootUntilRelease bool }

func (c *controls) releaseShoot() {
	if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
		c.suppressShootUntilRelease = false
	}
}

func (c *controls) read(camera rl.Camera2D, center rl.Vector2, facing simulation.Direction) protocol.Direction {
	c.releaseShoot()
	aim := mouseAim(camera, center, facing)
	direction := protocol.Direction{
		Shoot:    !c.suppressShootUntilRelease && rl.IsMouseButtonDown(rl.MouseLeftButton),
		Interact: rl.IsKeyPressed(rl.KeyE),
		Boost:    rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift),
		Dash:     rl.IsKeyPressed(rl.KeyLeftShift) || rl.IsKeyPressed(rl.KeyRightShift),
		AimX:     float64(aim.X), AimY: float64(aim.Y),
	}
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
	return direction
}

func playerCenter(p protocol.Player) rl.Vector2 {
	return rl.NewVector2(float32(p.X+p.Size/2), float32(p.Y+p.Size/2))
}

func mouseAim(camera rl.Camera2D, center rl.Vector2, fallback simulation.Direction) rl.Vector2 {
	target := rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
	delta := rl.Vector2Subtract(target, center)
	if rl.Vector2Length(delta) > 1 {
		return rl.Vector2Normalize(delta)
	}
	if fallback.X != 0 || fallback.Y != 0 {
		return rl.NewVector2(float32(fallback.X), float32(fallback.Y))
	}
	return rl.NewVector2(1, 0)
}

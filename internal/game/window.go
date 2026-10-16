package game

import (
	"image/color"

	"raylib-2d-sandbox/internal/display"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Window struct {
	width, height int32
	title         string
	bgColor       color.RGBA
}

func NewWindow(width, height int32, title string, bgColor color.RGBA) *Window {
	return &Window{width: width, height: height, title: title, bgColor: bgColor}
}

func (w *Window) initWindow() {
	rl.SetConfigFlags(rl.FlagWindowHidden)
	rl.InitWindow(w.width, w.height, w.title)
	rl.SetWindowMonitor(0)
	rl.ClearWindowState(rl.FlagWindowHidden)
	rl.SetWindowState(rl.FlagFullscreenMode)
	rl.SetWindowMonitor(0)
	rl.SetTargetFPS(60)
	rl.SetExitKey(rl.KeyNull)
	w.syncSize()
}

func (w *Window) syncSize() {
	w.width = int32(max(1, rl.GetScreenWidth()))
	w.height = int32(max(1, rl.GetScreenHeight()))
	display.Size(w.width, w.height)
	rl.SetMouseScale(1, 1)
}

func (w *Window) beginDrawing() {
	rl.BeginDrawing()
}

func (w *Window) endDrawing() {
	rl.EndDrawing()
}

func (w *Window) close() {
	rl.CloseWindow()
}

func (w *Window) getWindowCenter() rl.Vector2 {
	return rl.NewVector2(float32(w.width)/2, float32(w.height)/2)
}

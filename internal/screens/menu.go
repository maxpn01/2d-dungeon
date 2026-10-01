package screens

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Action int

const (
	ActionNone Action = iota
	ActionPlay
	ActionExit
	ActionMultiplayer
	ActionMenu
	ActionSettings
	ActionSaveSettings
)

type MenuButton struct {
	Text      string
	TextColor color.RGBA
	FontSize  int32

	BackgroundColor color.RGBA
	HoverColor      color.RGBA

	Bounds      rl.Rectangle
	BoundsColor color.RGBA

	Action Action
}

func NewMenuButton(text string, action Action) *MenuButton {
	return &MenuButton{
		Text:            text,
		TextColor:       rl.White,
		FontSize:        28,
		BackgroundColor: rl.Black,
		HoverColor:      rl.DarkGray,
		Bounds:          rl.NewRectangle(0, 0, 200, 60),
		BoundsColor:     rl.White,
		Action:          action,
	}

}

func (btn *MenuButton) Draw() Action {
	hovered := rl.CheckCollisionPointRec(
		rl.GetMousePosition(),
		btn.Bounds,
	)

	bgColor := btn.BackgroundColor

	if hovered {
		bgColor = btn.HoverColor
	}

	rl.DrawRectangleRec(btn.Bounds, bgColor)
	rl.DrawRectangleLinesEx(btn.Bounds, 2, btn.BoundsColor)

	textWidth := float32(rl.MeasureText(btn.Text, btn.FontSize))
	textX := btn.Bounds.X + (btn.Bounds.Width-textWidth)/2
	textY := btn.Bounds.Y + (btn.Bounds.Height-float32(btn.FontSize))/2

	rl.DrawText(btn.Text, int32(textX), int32(textY), btn.FontSize, btn.TextColor)

	if hovered && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return btn.Action
	}

	return ActionNone
}

type Menu struct {
	Title           string
	Options         []*MenuButton
	Gap             float32
	BackgroundColor color.RGBA
}

func NewMenu(options []*MenuButton, gap float32, bgColor color.RGBA) *Menu {
	return &Menu{
		Options:         options,
		Gap:             gap,
		BackgroundColor: bgColor,
	}
}

func (m *Menu) Draw() Action {
	rl.DrawRectangle(
		0, 0,
		int32(rl.GetScreenWidth()),
		int32(rl.GetScreenHeight()),
		m.BackgroundColor,
	)

	if m.Title != "" {
		cx := int32(rl.GetScreenWidth()) / 2
		cy := int32(rl.GetScreenHeight()) / 2
		pixelSize := max(int32(1), min(int32(rl.GetScreenWidth())/max(64, int32(len(m.Title)*6+11)), int32(rl.GetScreenHeight())/64))
		titleY := cy - 10*pixelSize

		rl.DrawRectangle(cx-2*pixelSize, titleY-8*pixelSize, 4*pixelSize, 4*pixelSize, rl.Red)
		drawPixelTitle(m.Title, cx, titleY, pixelSize, rl.RayWhite)

		return m.drawOptions(float32(cy + 5*pixelSize))
	}

	var totalHeight float32
	for i, btn := range m.Options {
		totalHeight += btn.Bounds.Height
		if i > 0 {
			totalHeight += m.Gap
		}
	}

	y := (float32(rl.GetScreenHeight()) - totalHeight) / 2

	return m.drawOptions(y)
}

func (m *Menu) drawOptions(y float32) Action {
	action := ActionNone

	for _, btn := range m.Options {
		btn.Bounds.X = (float32(rl.GetScreenWidth()) - btn.Bounds.Width) / 2
		btn.Bounds.Y = y

		if clickedAction := btn.Draw(); clickedAction != ActionNone {
			action = clickedAction
		}

		y += btn.Bounds.Height + m.Gap
	}

	return action
}

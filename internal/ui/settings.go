package ui

import (
	"strings"

	"2d-dungeon/internal/appearance"
	"2d-dungeon/internal/display"
	"2d-dungeon/internal/preferences"
	"2d-dungeon/internal/rendering/primitives"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Settings struct {
	preferences.Profile
	character  appearance.Character
	name       string
	focused    bool
	errorText  string
	saveButton *MenuButton
	backButton *MenuButton
}

func NewSettings() *Settings {
	return &Settings{
		Profile: preferences.Load(), saveButton: NewMenuButton("save", ActionSaveSettings),
		backButton: NewMenuButton("cancel", ActionMenu),
	}
}

func (s *Settings) Open() {
	s.name = s.PlayerName
	s.character = appearance.Normalize(s.Character)
	s.focused = true
	s.errorText = ""
	for rl.GetCharPressed() != 0 {
	}
}

func (s *Settings) save() bool {
	profile := preferences.Profile{PlayerName: preferences.NormalizePlayerName(s.name), Character: appearance.Normalize(s.character)}
	err := profile.Save()
	if err != nil {
		s.errorText = "Could not save settings. Please try again."
		return false
	}
	s.Profile = profile
	return true
}

func (s *Settings) Draw() Action {
	cx := float32(display.Width()) / 2
	top := float32(display.Height())/2 - 230
	primitives.DrawCenteredText("settings", int32(top), 40, rl.RayWhite)
	for _, character := range []appearance.Character{appearance.Square} {
		card := rl.NewRectangle(cx-75, top+65, 150, 115)
		enabled := appearance.Enabled(character)
		if enabled && rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), card) {
			s.character = character
		}
		tint := rl.DarkGray
		if s.character == character {
			tint = primitives.CharacterColor(character)
		}
		rl.DrawRectangleLinesEx(card, 2, tint)
		shapeTint, textTint := primitives.CharacterColor(character), rl.RayWhite
		if !enabled {
			shapeTint, textTint = rl.DarkGray, rl.Gray
		}
		primitives.DrawCharacterShape(character, rl.NewRectangle(card.X+55, card.Y+15, 40, 40), shapeTint)
		primitives.DrawText(string(character), int32(card.X+75)-primitives.MeasureText(string(character), 20)/2, int32(card.Y+75), 20, textTint)
		if !enabled {
			primitives.DrawText("disabled", int32(card.X+75)-primitives.MeasureText("disabled", 14)/2, int32(card.Y+96), 14, rl.Gray)
		}
	}
	bounds := rl.NewRectangle(cx-190, top+255, 380, 52)
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.focused = rl.CheckCollisionPointRec(rl.GetMousePosition(), bounds)
	}
	for character := rl.GetCharPressed(); character != 0; character = rl.GetCharPressed() {
		if s.focused && character >= 32 && character <= 126 && len(s.name) < preferences.MaxPlayerNameLength {
			s.name += string(rune(character))
		}
	}
	if s.focused {
		if (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) && rl.IsKeyPressed(rl.KeyA) {
			s.name = ""
		}
		if (rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace)) && len(s.name) > 0 {
			s.name = s.name[:len(s.name)-1]
		}
	}
	primitives.DrawText("player name", int32(bounds.X), int32(bounds.Y)-30, 22, rl.RayWhite)
	tint := rl.Gray
	if s.focused {
		tint = rl.RayWhite
	}
	rl.DrawRectangleLinesEx(bounds, 2, tint)
	fontSize := int32(24)
	for primitives.MeasureText(s.name, fontSize) > int32(bounds.Width)-28 && fontSize > 12 {
		fontSize--
	}
	primitives.DrawText(s.name, int32(bounds.X)+12, int32(bounds.Y)+(int32(bounds.Height)-fontSize)/2, fontSize, rl.RayWhite)
	if s.focused && int(rl.GetTime()*2)%2 == 0 {
		x := int32(bounds.X) + 12 + primitives.MeasureText(s.name, fontSize)
		rl.DrawRectangle(x, int32(bounds.Y)+12, 2, 28, rl.RayWhite)
	}
	primitives.DrawCenteredText("Up to 20 characters | Ctrl+A to clear", int32(top)+315, 18, rl.LightGray)
	s.saveButton.Bounds = rl.NewRectangle(cx-190, top+360, 180, 60)
	s.backButton.Bounds = rl.NewRectangle(cx+10, top+360, 180, 60)
	save := s.saveButton.Draw()
	back := s.backButton.Draw()
	if s.errorText != "" {
		primitives.DrawCenteredText(s.errorText, int32(top)+435, 18, rl.Red)
	}
	if back != ActionNone {
		return back
	}
	if save != ActionNone || rl.IsKeyPressed(rl.KeyEnter) {
		if strings.TrimSpace(s.name) == "" {
			s.errorText = "Please enter a player name."
		} else if s.save() {
			return ActionSaveSettings
		}
	}
	return ActionNone
}

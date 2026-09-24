package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Settings struct {
	Character  simulation.Character `json:"character"`
	character  simulation.Character
	PlayerName string `json:"player_name"`
	name       string
	focused    bool
	errorText  string
	saveButton *MenuButton
	backButton *MenuButton
}

func newSettings() *Settings {
	s := &Settings{
		PlayerName: "Player", Character: simulation.Square,
		saveButton: NewMenuButton("save", ActionSaveSettings),
		backButton: NewMenuButton("cancel", ActionMenu),
	}
	if path, err := settingsPath(); err == nil {
		if data, err := os.ReadFile(path); err == nil {
			var saved struct {
				PlayerName string               `json:"player_name"`
				Character  simulation.Character `json:"character"`
			}
			if json.Unmarshal(data, &saved) == nil {
				s.PlayerName = protocol.NormalizePlayerName(saved.PlayerName)
				s.Character = simulation.NormalizeCharacter(saved.Character)
			}
		}
	}
	return s
}

func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "raylib-2d-sandbox", "settings.json"), nil
}

func (s *Settings) open() {
	s.name = s.PlayerName
	s.character = s.Character
	s.focused = true
	s.errorText = ""
	for rl.GetCharPressed() != 0 {
	}
}

func (s *Settings) save() bool {
	name := protocol.NormalizePlayerName(s.name)
	path, err := settingsPath()
	if err == nil {
		err = os.MkdirAll(filepath.Dir(path), 0755)
	}
	if err == nil {
		data, _ := json.Marshal(struct {
			PlayerName string               `json:"player_name"`
			Character  simulation.Character `json:"character"`
		}{name, simulation.NormalizeCharacter(s.character)})
		err = os.WriteFile(path, data, 0644)
	}
	if err != nil {
		s.errorText = "Could not save settings. Please try again."
		return false
	}
	s.PlayerName = name
	s.Character = simulation.NormalizeCharacter(s.character)
	return true
}

func (s *Settings) draw() Action {
	cx := float32(rl.GetScreenWidth()) / 2
	top := float32(rl.GetScreenHeight())/2 - 230
	drawCenteredText("settings", int32(top), 40, rl.RayWhite)
	for i, character := range []simulation.Character{simulation.Square, simulation.Triangle, simulation.Circle} {
		card := rl.NewRectangle(cx-240+float32(i)*165, top+65, 150, 115)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(rl.GetMousePosition(), card) {
			s.character = character
		}
		tint := rl.DarkGray
		if s.character == character {
			tint = characterColor(character)
		}
		rl.DrawRectangleLinesEx(card, 2, tint)
		drawCharacterShape(character, rl.NewRectangle(card.X+55, card.Y+15, 40, 40), characterColor(character))
		drawText(string(character), int32(card.X+75)-measureText(string(character), 20)/2, int32(card.Y+75), 20, rl.RayWhite)
	}
	bounds := rl.NewRectangle(cx-190, top+255, 380, 52)
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.focused = rl.CheckCollisionPointRec(rl.GetMousePosition(), bounds)
	}
	for character := rl.GetCharPressed(); character != 0; character = rl.GetCharPressed() {
		if s.focused && character >= 32 && character <= 126 && len(s.name) < protocol.MaxPlayerNameLength {
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
	drawText("player name", int32(bounds.X), int32(bounds.Y)-30, 22, rl.RayWhite)
	tint := rl.Gray
	if s.focused {
		tint = rl.RayWhite
	}
	rl.DrawRectangleLinesEx(bounds, 2, tint)
	fontSize := int32(24)
	for measureText(s.name, fontSize) > int32(bounds.Width)-28 && fontSize > 12 {
		fontSize--
	}
	drawText(s.name, int32(bounds.X)+12, int32(bounds.Y)+(int32(bounds.Height)-fontSize)/2, fontSize, rl.RayWhite)
	if s.focused && int(rl.GetTime()*2)%2 == 0 {
		x := int32(bounds.X) + 12 + measureText(s.name, fontSize)
		rl.DrawRectangle(x, int32(bounds.Y)+12, 2, 28, rl.RayWhite)
	}
	drawCenteredText("Up to 20 characters | Ctrl+A to clear", int32(top)+315, 18, rl.LightGray)
	s.saveButton.Bounds = rl.NewRectangle(cx-190, top+360, 180, 60)
	s.backButton.Bounds = rl.NewRectangle(cx+10, top+360, 180, 60)
	save := s.saveButton.draw()
	back := s.backButton.draw()
	if s.errorText != "" {
		drawCenteredText(s.errorText, int32(top)+435, 18, rl.Red)
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

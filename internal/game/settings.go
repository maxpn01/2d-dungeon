package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"raylib-2d-sandbox/internal/protocol"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Settings struct {
	PlayerName string `json:"player_name"`
	name       string
	focused    bool
	errorText  string
	saveButton *MenuButton
	backButton *MenuButton
}

func newSettings() *Settings {
	s := &Settings{
		PlayerName: "Player",
		saveButton: NewMenuButton("save", ActionSaveSettings),
		backButton: NewMenuButton("cancel", ActionMenu),
	}
	if path, err := settingsPath(); err == nil {
		if data, err := os.ReadFile(path); err == nil {
			var saved struct {
				PlayerName string `json:"player_name"`
			}
			if json.Unmarshal(data, &saved) == nil {
				s.PlayerName = protocol.NormalizePlayerName(saved.PlayerName)
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
			PlayerName string `json:"player_name"`
		}{name})
		err = os.WriteFile(path, data, 0644)
	}
	if err != nil {
		s.errorText = "Could not save settings. Please try again."
		return false
	}
	s.PlayerName = name
	return true
}

func (s *Settings) draw() Action {
	cx := float32(rl.GetScreenWidth()) / 2
	top := float32(rl.GetScreenHeight())/2 - 230
	drawCenteredText("settings", int32(top), 40, rl.RayWhite)
	rl.DrawRectangle(int32(cx)-30, int32(top)+70, 60, 60, rl.Red)
	drawCenteredText("avatar", int32(top)+145, 20, rl.LightGray)
	bounds := rl.NewRectangle(cx-190, top+215, 380, 52)
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
	rl.DrawText("player name", int32(bounds.X), int32(bounds.Y)-30, 22, rl.RayWhite)
	tint := rl.Gray
	if s.focused {
		tint = rl.RayWhite
	}
	rl.DrawRectangleLinesEx(bounds, 2, tint)
	fontSize := int32(24)
	for rl.MeasureText(s.name, fontSize) > int32(bounds.Width)-28 && fontSize > 12 {
		fontSize--
	}
	rl.DrawText(s.name, int32(bounds.X)+12, int32(bounds.Y)+(int32(bounds.Height)-fontSize)/2, fontSize, rl.RayWhite)
	if s.focused && int(rl.GetTime()*2)%2 == 0 {
		x := int32(bounds.X) + 12 + rl.MeasureText(s.name, fontSize)
		rl.DrawRectangle(x, int32(bounds.Y)+12, 2, 28, rl.RayWhite)
	}
	drawCenteredText("Up to 20 characters | Ctrl+A to clear", int32(top)+280, 18, rl.LightGray)
	s.saveButton.Bounds = rl.NewRectangle(cx-190, top+330, 180, 60)
	s.backButton.Bounds = rl.NewRectangle(cx+10, top+330, 180, 60)
	save := s.saveButton.draw()
	back := s.backButton.draw()
	if s.errorText != "" {
		drawCenteredText(s.errorText, int32(top)+410, 18, rl.Red)
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

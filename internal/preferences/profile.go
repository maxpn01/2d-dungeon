package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"

	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/simulation"
)

type Profile struct {
	PlayerName string               `json:"player_name"`
	Character  simulation.Character `json:"character"`
}

func Load() Profile {
	profile := Profile{PlayerName: "Player", Character: simulation.Square}
	path, err := settingsPath()
	if err != nil {
		return profile
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return profile
	}
	var saved Profile
	if json.Unmarshal(data, &saved) != nil {
		return profile
	}
	saved.normalize()
	return saved
}

func (p Profile) Save() error {
	p.normalize()
	path, err := settingsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (p *Profile) normalize() {
	p.PlayerName = protocol.NormalizePlayerName(p.PlayerName)
	p.Character = simulation.NormalizeCharacter(p.Character)
}

func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "raylib-2d-sandbox", "settings.json"), nil
}

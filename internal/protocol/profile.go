package protocol

import "strings"

const MaxPlayerNameLength = 20

func NormalizePlayerName(name string) string {
	var normalized strings.Builder
	for _, character := range name {
		if character >= 32 && character <= 126 {
			normalized.WriteRune(character)
		}
		if normalized.Len() == MaxPlayerNameLength {
			break
		}
	}
	name = strings.TrimSpace(normalized.String())
	if name == "" {
		return "Player"
	}
	return name
}

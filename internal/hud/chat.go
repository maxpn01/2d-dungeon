package hud

import (
	"fmt"
	"time"

	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
	"raylib-2d-sandbox/internal/rendering/primitives"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	chatVisibleLines = parameters.ChatVisibleLines
	chatFadeSeconds  = parameters.ChatFadeSeconds
	chatFontSize     = parameters.ChatFontSize
	chatLineHeight   = parameters.ChatLineHeight
	chatMargin       = parameters.ChatMargin
)

type ChatBox struct {
	draft      string
	Open       bool
	openingKey bool
}

func (c *ChatBox) Close() {
	c.Open = false
	c.openingKey = false
	c.draft = ""
}

func (c *ChatBox) Update(ready, dead bool, errorText string, send func(string)) {
	if c.Open && (!ready || errorText != "" || dead) {
		c.Close()
		return
	}
	if !c.Open {
		if ready && errorText == "" && !dead && rl.IsWindowFocused() && rl.IsKeyPressed(rl.KeyT) {
			c.Open = true
			c.openingKey = true
			c.draft = ""
			for rl.GetCharPressed() != 0 {
			}
		}
		return
	}
	for character := rl.GetCharPressed(); character != 0; character = rl.GetCharPressed() {
		if c.openingKey && (character == 't' || character == 'T') {
			continue
		}
		if character >= 32 && character <= 126 && len(c.draft) < protocol.MaxChatMessageLength {
			c.draft += string(rune(character))
		}
	}
	if !rl.IsKeyDown(rl.KeyT) {
		c.openingKey = false
	}
	if (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) && rl.IsKeyPressed(rl.KeyA) {
		c.draft = ""
	}
	if (rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace)) && len(c.draft) > 0 {
		c.draft = c.draft[:len(c.draft)-1]
	}
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
		send(c.draft)
		c.Close()
	}
}

func (c *ChatBox) Draw(width, height int32, entries []protocol.ChatEntry) {
	if len(entries) > chatVisibleLines {
		entries = entries[len(entries)-chatVisibleLines:]
	}
	inputTop := float32(height - chatMargin - 34)
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		alpha := float32(1)
		if !c.Open {
			remaining := chatFadeSeconds - time.Since(entry.At).Seconds()
			if remaining <= 0 {
				continue
			}
			if remaining < 1 {
				alpha = float32(remaining)
			}
		}
		y := int32(inputTop) - chatLineHeight*int32(len(entries)-i) - 8
		line := fmt.Sprintf("<%s> %s", entry.Name, entry.Text)
		tint := rl.RayWhite
		if entry.Mine {
			tint = rl.SkyBlue
		}
		rl.DrawRectangle(chatMargin-8, y-4, primitives.MeasureText(line, chatFontSize)+16, chatLineHeight, rl.Fade(rl.Black, 0.5*alpha))
		primitives.DrawText(line, chatMargin, y, chatFontSize, rl.Fade(tint, alpha))
	}
	if !c.Open {
		return
	}
	bounds := rl.NewRectangle(chatMargin, inputTop, float32(width)-2*chatMargin, 34)
	rl.DrawRectangleRec(bounds, rl.Fade(rl.Black, 0.7))
	rl.DrawRectangleLinesEx(bounds, 2, rl.RayWhite)
	primitives.DrawText(c.draft, int32(bounds.X)+10, int32(bounds.Y)+8, chatFontSize, rl.RayWhite)
	if int(rl.GetTime()*2)%2 == 0 {
		rl.DrawRectangle(int32(bounds.X)+10+primitives.MeasureText(c.draft, chatFontSize), int32(bounds.Y)+7, 2, chatFontSize+2, rl.RayWhite)
	}
	primitives.DrawText("Enter to send | Esc to cancel", chatMargin, int32(inputTop)-chatLineHeight*int32(chatVisibleLines)-34, 16, rl.LightGray)
}

package game

import (
	"fmt"
	"time"

	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/protocol"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	chatVisibleLines = 8
	chatFadeSeconds  = 10
	chatFontSize     = 18
	chatLineHeight   = 24
	chatMargin       = 24
)

type chatBox struct {
	draft string
	open  bool
}

func (c *chatBox) close() {
	c.open = false
	c.draft = ""
}

func (c *chatBox) update(m *multiplayer) {
	if c.open && (!m.view.Ready || m.errorText != "" || m.view.Local.Dead) {
		c.close()
		return
	}
	if !c.open {
		if m.view.Ready && m.errorText == "" && !m.view.Local.Dead && rl.IsWindowFocused() && rl.IsKeyPressed(rl.KeyT) {
			c.open = true
			c.draft = ""
			for rl.GetCharPressed() != 0 {
			}
		}
		return
	}
	for character := rl.GetCharPressed(); character != 0; character = rl.GetCharPressed() {
		if character >= 32 && character <= 126 && len(c.draft) < protocol.MaxChatMessageLength {
			c.draft += string(rune(character))
		}
	}
	if (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) && rl.IsKeyPressed(rl.KeyA) {
		c.draft = ""
	}
	if (rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace)) && len(c.draft) > 0 {
		c.draft = c.draft[:len(c.draft)-1]
	}
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
		m.session.SendChat(c.draft)
		c.close()
	}
}

func (c *chatBox) draw(window *Window, entries []client.ChatEntry) {
	if len(entries) > chatVisibleLines {
		entries = entries[len(entries)-chatVisibleLines:]
	}
	inputTop := float32(window.height - chatMargin - 34)
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		alpha := float32(1)
		if !c.open {
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
		rl.DrawRectangle(chatMargin-8, y-4, rl.MeasureText(line, chatFontSize)+16, chatLineHeight, rl.Fade(rl.Black, 0.5*alpha))
		rl.DrawText(line, chatMargin, y, chatFontSize, rl.Fade(tint, alpha))
	}
	if !c.open {
		return
	}
	bounds := rl.NewRectangle(chatMargin, inputTop, float32(window.width)-2*chatMargin, 34)
	rl.DrawRectangleRec(bounds, rl.Fade(rl.Black, 0.7))
	rl.DrawRectangleLinesEx(bounds, 2, rl.RayWhite)
	rl.DrawText(c.draft, int32(bounds.X)+10, int32(bounds.Y)+8, chatFontSize, rl.RayWhite)
	if int(rl.GetTime()*2)%2 == 0 {
		rl.DrawRectangle(int32(bounds.X)+10+rl.MeasureText(c.draft, chatFontSize), int32(bounds.Y)+7, 2, chatFontSize+2, rl.RayWhite)
	}
	rl.DrawText("Enter to send | Esc to cancel", chatMargin, int32(inputTop)-chatLineHeight*int32(chatVisibleLines)-34, 16, rl.LightGray)
}

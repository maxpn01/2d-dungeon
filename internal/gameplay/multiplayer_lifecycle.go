package gameplay

import (
	"raylib-2d-sandbox/internal/client"
	"raylib-2d-sandbox/internal/protocol"
)

func (m *Multiplayer) Close()        { m.session.Close() }
func (m *Multiplayer) Started() bool { return m.view.MatchStarted }
func (m *Multiplayer) StartGame()    { m.session.StartGame() }
func (m *Multiplayer) Ready() bool   { return m.view.Ready }
func (m *Multiplayer) Dead() bool    { return m.view.Local.Dead }
func (m *Multiplayer) Error() string { return m.errorText }
func (m *Multiplayer) PerksOpen() bool {
	return !m.ResultsOpen() && m.view.Local.ActiveStatue != "" && !m.perks.Closing
}
func (m *Multiplayer) ChatOpen() bool { return m.chat.Open }
func (m *Multiplayer) CloseChat()     { m.chat.Close() }
func (m *Multiplayer) CanOpenMap() bool {
	return m.Ready() && !m.ResultsOpen() && m.view.MatchStarted && !m.Dead() && m.Error() == "" && !m.ChatOpen() && !m.StatsOpen() && m.view.Local.ActiveStatue == ""
}
func (m *Multiplayer) ChatHistory() []protocol.ChatEntry        { return m.session.Chat() }
func (m *Multiplayer) RestoreChat(history []protocol.ChatEntry) { m.session.RestoreChat(history) }
func (m *Multiplayer) Respawn() {
	m.CloseStats()
	m.session.Respawn()
	m.motion = client.Motion{}
	m.controls.suppressShootUntilRelease = true
}

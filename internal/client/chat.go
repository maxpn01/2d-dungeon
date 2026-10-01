package client

import (
	"raylib-2d-sandbox/internal/parameters"
	"raylib-2d-sandbox/internal/protocol"
)

const maxChatEntries = parameters.MaxChatEntries

func (s *Session) SendChat(text string) {
	text = protocol.NormalizeChatMessage(text)
	if text == "" {
		return
	}
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	if len(s.outgoingChat) < 8 {
		s.outgoingChat = append(s.outgoingChat, text)
	}
}

func (s *Session) Chat() []protocol.ChatEntry {
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	return append([]protocol.ChatEntry(nil), s.chat...)
}

func (s *Session) RestoreChat(entries []protocol.ChatEntry) {
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	history := append([]protocol.ChatEntry(nil), entries...)
	s.chat = append(history, s.chat...)
	if len(s.chat) > maxChatEntries {
		s.chat = append([]protocol.ChatEntry(nil), s.chat[len(s.chat)-maxChatEntries:]...)
	}
}

func (s *Session) appendChat(entry protocol.ChatEntry) {
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	s.chat = append(s.chat, entry)
	if len(s.chat) > maxChatEntries {
		s.chat = append([]protocol.ChatEntry(nil), s.chat[len(s.chat)-maxChatEntries:]...)
	}
}

func (s *Session) takeOutgoingChat() []string {
	s.chatMu.Lock()
	defer s.chatMu.Unlock()
	pending := s.outgoingChat
	s.outgoingChat = nil
	return pending
}

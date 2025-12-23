package mstore

import (
	"sync"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type Message struct {
	Info    types.MessageInfo
	Content *waE2E.Message
}

type MessageStore struct {
	mu     sync.RWMutex
	msgMap map[types.JID][]Message
}

func NewMessageStore() *MessageStore {
	return &MessageStore{
		msgMap: make(map[types.JID][]Message),
	}
}

func (ms *MessageStore) ProcessMessageEvent(msg *events.Message) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	chat := msg.Info.Chat
	ms.msgMap[chat] = append(ms.msgMap[chat], Message{
		Info:    msg.Info,
		Content: msg.Message,
	})
}

func (ms *MessageStore) GetMessages(jid types.JID) []Message {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.msgMap[jid]
}

func (ms *MessageStore) GetLatestMessage(jid types.JID) *Message {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	messages, exists := ms.msgMap[jid]
	if !exists || len(messages) == 0 {
		return nil
	}
	return &messages[len(messages)-1]
}

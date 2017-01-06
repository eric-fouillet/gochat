package gochatutil

import (
	"sync"

	"github.com/ericfouillet/gochat"
)

// MsgPool implements a simple message pool
type MsgPool struct {
	pool sync.Pool
}

// NewPool creates a new message pool
func NewPool() *MsgPool {
	var p MsgPool
	p.pool.New = NewMsg
	return &p
}

func NewMsg() interface{} {
	return new(gochat.ChatMessage)
}

// Get gets a message from the pool
func (p *MsgPool) Get() *gochat.ChatMessage {
	return p.pool.Get().(*gochat.ChatMessage)
}

// Rel releases a message to the pool
func (p *MsgPool) Rel(m *gochat.ChatMessage) {
	p.pool.Put(m)
}

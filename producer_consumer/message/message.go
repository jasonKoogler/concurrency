package message

import (
	"fmt"
	"sync"
	"time"
)

type message struct {
	data any
}

type Messenger interface {
	Produce() <-chan message
	Consume() <-chan message
}

type messenger struct {
	msgch chan message
	wg    *sync.WaitGroup
}

func NewMessenger(bufferSize int, wg *sync.WaitGroup) *messenger {
	return &messenger{
		msgch: make(chan message, bufferSize),
		wg:    wg,
	}
}

func (m *messenger) Produce(max int) <-chan message {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for i := 0; i < max; i++ {
			m.msgch <- message{data: fmt.Sprintf("Message %d", i)}
			time.Sleep(1 * time.Second)
		}
		close(m.msgch)
	}()
	return m.msgch
}

func (m *messenger) Consume() <-chan message {
	m.wg.Add(1)
	out := make(chan message)
	go func() {
		defer m.wg.Done()
		defer close(out)
		for msg := range m.msgch {
			fmt.Printf("Consumed: %s\n", msg.data)
			time.Sleep(2 * time.Second)
		}
	}()
	return out
}

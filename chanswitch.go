package chanswitch

import (
	"context"
	"fmt"
	"sync"
)

/*
select {

case <- test.Is(100).Chan():

case <- test.Is(200).Chan():

case <- test.Is(300).Chan():

}
*/

type Channels struct {
	open chan struct{}
	once chan struct{}
}

type BroadAny struct {
	filters map[any]*Channels
	Value   any
	old     any
	fl      sync.RWMutex
}

var active = struct{}{}

func New() *BroadAny {
	b := &BroadAny{
		filters: make(map[any]*Channels),
		fl:      sync.RWMutex{},
	}

	return b
}

// set filed value
func (b *BroadAny) Set(v any) {
	ch := b.make(v)

	b.old = b.Value
	b.Value = v
	activeChan(ch.open)
	activeChan(ch.once)

	// deactive channel got other value
	for k, ch := range b.filters {
		if v != k {
			closeChan(ch.open)
			closeChan(ch.once)
		}
	}

}

// thread safe wait until this field change to true
func (b *BroadAny) WaitFor(ctx context.Context, v any) {
	ch := b.make(v)

	defer func() {
		if b.Value == v {
			if ch != nil {
				activeChan(ch.open)
			}
		}
	}()

	select {
	case <-ctx.Done():
		fmt.Printf("context for %v is timeout\n", v)
	case <-ch.open:
		fmt.Printf("wait for %v done\n", v)
	}
}

// check whether this field is true or not
func (b *BroadAny) On(v any) <-chan struct{} { // Running Reproducible
	ch := b.make(v)

	defer func() {
		if b.Value == v {
			if ch != nil {
				activeChan(ch.open)
			}
		}
	}()

	return ch.open
}

func (b *BroadAny) Once(v any) <-chan struct{} { // Running Once
	ch := b.make(v)
	return ch.once
}

func (b *BroadAny) read(v any) *Channels {
	b.fl.RLock()
	defer b.fl.RUnlock()

	ch, ok := b.filters[v]
	if !ok {
		return nil
	}
	return ch
}

func (b *BroadAny) set(v any, ch *Channels) {
	b.fl.Lock()
	defer b.fl.Unlock()

	b.filters[v] = ch
}

func (b *BroadAny) make(v any) *Channels {
	if chs := b.read(v); chs == nil {
		ch := &Channels{
			open: make(chan struct{}, 1),
			once: make(chan struct{}, 1),
		}
		b.set(v, ch)
		return ch
	} else {
		return chs
	}
}

func activeChan(c chan struct{}) {
	if len(c) == 0 {
		c <- struct{}{}
	}
}

func closeChan(c chan struct{}) {
	select {
	case <-c:
	default:
	}
}

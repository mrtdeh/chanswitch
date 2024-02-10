package chanswitch

import (
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
	sl      sync.Mutex
	fl      sync.RWMutex
}

var active = struct{}{}

func New() *BroadAny {
	b := &BroadAny{
		filters: make(map[any]*Channels),
		sl:      sync.Mutex{},
		fl:      sync.RWMutex{},
	}

	return b
}

// set filed value
func (b *BroadAny) Switch(v any) {
	b.sl.Lock()
	defer b.sl.Unlock()
	// make sure channels are created
	b.make(v)

	// active channel for this value
	if ch := b.read(v); ch != nil {
		b.old = b.Value
		b.Value = v
		activeChan(ch.open)
		activeChan(ch.once)

	} else {
		panic("invalid value : filter value is not exist")
	}
	// deactive channel got other value
	for k, ch := range b.filters {
		if v != k {
			closeChan(ch.open)
			closeChan(ch.once)
		}
	}

}

// check whether this field is true or not
func (b *BroadAny) On(v any) <-chan struct{} { // Running Reproducible
	// make sure channels are created
	b.make(v)

	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("on error : value %v not found", v))
	}

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
	// make sure channels are created
	b.make(v)

	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("once error : value %v not found", v))
	}

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

func (b *BroadAny) make(v any) {
	if chs := b.read(v); chs == nil {
		b.set(v, &Channels{
			open: make(chan struct{}, 1),
			once: make(chan struct{}, 1),
		})
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

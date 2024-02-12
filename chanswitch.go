package chanswitch

import (
	"context"
	"fmt"
)

type Channels struct {
	open chan struct{}
	once chan struct{}
}

type ChanSwitch struct {
	filters map[any]*Channels
	Value   any
}

var active = struct{}{}

func New(vals ...any) *ChanSwitch {
	b := &ChanSwitch{
		filters: make(map[any]*Channels),
	}

	for _, v := range vals {
		b.Make(v)
	}

	return b
}

// set filed value
func (b *ChanSwitch) Set(v any) {
	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("value not set : %v", v))
	}

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
func (b *ChanSwitch) WaitFor(ctx context.Context, v any) {
	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("value not set : %v", v))
	}

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
func (b *ChanSwitch) On(v any) <-chan struct{} { // Running Reproducible
	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("value not set : %v", v))
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

func (b *ChanSwitch) Once(v any) <-chan struct{} { // Running Once
	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("value not set : %v", v))
	}
	return ch.once
}

func (b *ChanSwitch) read(v any) *Channels {
	return b.filters[v]
}

func (b *ChanSwitch) set(v any, ch *Channels) {
	b.filters[v] = ch
}

func (b *ChanSwitch) Make(v any) *Channels {
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

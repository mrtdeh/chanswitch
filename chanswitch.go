package chanswitch

import (
	"context"
	"fmt"
	"sync"
)

type Channels struct {
	stoped bool
	open   chan struct{}
	once   chan struct{}
	first  chan struct{}
	stop   chan struct{}
}

type ChanSwitch struct {
	filters map[any]*Channels
	val     any
	l       sync.Mutex
}

func New(vals ...any) *ChanSwitch {
	b := &ChanSwitch{
		filters: make(map[any]*Channels),
		l:       sync.Mutex{},
	}

	for _, v := range vals {
		b.Make(v)
	}

	return b
}

func NewBool() *ChanSwitch {
	b := &ChanSwitch{
		filters: make(map[any]*Channels),
		l:       sync.Mutex{},
	}

	b.Make(true)
	b.Make(false)

	return b
}

func (b *ChanSwitch) Make(v any) *Channels {
	if chs := b.read(v); chs == nil {
		ch := &Channels{
			open:  make(chan struct{}, 1),
			first: make(chan struct{}, 1),
			once:  make(chan struct{}, 1),
			stop:  make(chan struct{}, 1),
		}
		b.set(v, ch)
		return ch
	} else {
		return chs
	}
}

// set filed value
func (b *ChanSwitch) Set(v any) {
	b.l.Lock()
	defer b.l.Unlock()

	// get filter by value
	ch := b.filters[v]
	// update val
	b.val = v
	// reset stoped bool
	ch.stoped = false
	// reset stop channel
	closeChan(ch.stop)
	// reset first channel
	closeChan(ch.first)
	// reset once channel
	activeChan(ch.once)
	// goroutine for repeating open channel
	go func(c *Channels) {
		for {
			select {
			// on call stop channel
			case <-c.stop:
				// set stoped to true
				c.stoped = true
				return
				// re-val open channel
			case c.open <- struct{}{}:
			}
		}
	}(ch)
	// deactive channel got other value
	for k, c := range b.filters {
		if v != k {
			if !c.stoped && len(c.stop) == 0 {
				// active stop channel of old filter
				activeChan(c.stop)
				// close once channel of old filter
				closeChan(c.once)
				// close open channel of old filter
				closeChan(c.open)
				// close first channel of old filter
				closeChan(c.first)
			}
		}
	}
	// active first channel
	activeChan(ch.first)
}

// thread safe wait until this field change to true
func (b *ChanSwitch) WaitFor(ctx context.Context, v any) {
	ch := b.filters[v]

	select {
	case <-ctx.Done():
	case <-ch.first:
	}

}

// check whether this field is true or not
func (b *ChanSwitch) On(v any) chan struct{} { // Running Reproducible

	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("value not set : %v", v))
	}

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

func (b *ChanSwitch) Value() any {
	return b.val
}

func (b *ChanSwitch) Values() []any {
	keys := make([]any, 0, len(b.filters))
	for k := range b.filters {
		keys = append(keys, k)
	}
	return keys
}

func (b *ChanSwitch) log(v any, format string, a ...any) {
	if b.val == v || v == "any" {
		fmt.Printf(format+"\n", a...)
	}
}

func activeChan(c chan struct{}) {
	select {
	case c <- struct{}{}:
	default:
	}
}

func closeChan(c chan struct{}) {
	select {
	case <-c:
	default:
	}
}

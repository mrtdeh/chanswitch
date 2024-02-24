package chanswitch

import (
	"context"
	"fmt"
	"sync"
)

type Channels struct {
	open chan struct{}
	once chan struct{}
}

type ChanSwitch struct {
	filters  map[any]*Channels
	val      any
	distract chan struct{}
	l        sync.Mutex
}

func (b *ChanSwitch) currentChan() chan struct{} {
	if c := b.filters[b.val]; c != nil {
		return c.open
	}

	empty := make(chan struct{})
	// defer close(empty)

	return empty
}

func repeater(b *ChanSwitch) {

	for {
		select {
		case b.currentChan() <- struct{}{}:
		case <-b.distract:
		}
	}

}
func New(vals ...any) *ChanSwitch {
	b := &ChanSwitch{
		filters:  make(map[any]*Channels),
		distract: make(chan struct{}, 1),
		l:        sync.Mutex{},
	}

	go repeater(b)

	for _, v := range vals {
		b.Make(v)
	}

	return b
}

func NewBool() *ChanSwitch {
	return New(true, false)
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

// set filed value
func (b *ChanSwitch) Set(v any) {
	fmt.Println("debug 1")
	b.l.Lock()
	defer b.l.Unlock()
	fmt.Println("debug 2")
	// get filter by value
	ch := b.filters[v]
	// update val
	b.val = v
	// distract repeated goroutines for change filter
	b.distract <- struct{}{}
	fmt.Println("debug 3")
	// reset once channel
	activeChan(ch.once)

	// deactive channel got other value
	for k, c := range b.filters {
		if v != k {
			if len(c.open) > 0 {
				// close once channel of old filter
				closeChan(c.once)
				// close open channel of old filter
				closeChan(c.open)
			}
		}
	}
}

// thread safe wait until this field change to true
func (b *ChanSwitch) WaitFor(ctx context.Context, v any) {
	ch := b.filters[v]

	select {
	case <-ctx.Done():
	case <-ch.open:
		activeChan(ch.open)
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

// func (b *ChanSwitch) log(v any, format string, a ...any) {
// 	if b.val == v || v == "any" {
// 		fmt.Printf(format+"\n", a...)
// 	}
// }

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

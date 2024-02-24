package chanswitch

import (
	"context"
	"fmt"
	"sync"
)

type Channels struct {
	open  chan struct{}
	once  chan struct{}
	first chan struct{}
}

type ChanSwitch struct {
	filters  map[any]*Channels
	val      any
	l        sync.Mutex
	distract chan struct{}
	a        int
	b        int
}

func (b *ChanSwitch) currentChan() chan struct{} {
	if c := b.filters[b.val]; c != nil {
		return c.open
	}
	return nil
}

func New(vals ...any) *ChanSwitch {
	b := &ChanSwitch{
		filters:  make(map[any]*Channels),
		l:        sync.Mutex{},
		distract: make(chan struct{}, 1),
	}

	go func() {
		for {
			select {
			case b.currentChan() <- struct{}{}:
			case <-b.distract:
				b.b++
				// fmt.Println("distract for change to ", b.val)
			}
		}
	}()

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
	b.a++
	// get filter by value
	ch := b.filters[v]
	// update val
	b.val = v
	// distract repeated goroutines for change filter
	b.distract <- struct{}{}
	// reset first channel
	closeChan(ch.first)
	// reset once channel
	activeChan(ch.once)

	// deactive channel got other value
	for k, c := range b.filters {
		if v != k {
			if len(c.open) > 0 {
				closeChan(c.once)
				// close open channel of old filter
				closeChan(c.open)
				// close first channel of old filter
				closeChan(ch.first)
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
		activeChan(ch.first)
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

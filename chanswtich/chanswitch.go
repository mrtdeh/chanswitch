package chanswitch

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

type Channels struct {
	open chan struct{}
	once chan struct{}
}

type ChanSwitch struct {
	filters  map[any]*Channels
	val      any
	cond     *sync.Cond
	distract chan struct{}
	l        *sync.RWMutex
}

func (b *ChanSwitch) currentChan() chan struct{} {
	if c := b.read(b.val); c != nil {
		return c.open
	}

	empty := make(chan struct{})

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
		l:        &sync.RWMutex{},
		cond:     sync.NewCond(&sync.Mutex{}), //&sync.Cond{},
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
		b.store(v, ch)
		return ch
	} else {
		return chs
	}
}

// set filed value
func (b *ChanSwitch) Set(v any) {

	if b.val != nil {
		oldch := b.read(b.val)
		cleanChan(oldch.once)
		cleanChan(oldch.open)
	}

	b.val = v
	activeChan(b.distract, struct{}{})

	ch := b.read(v)
	activeChan(ch.once)

	b.cond.Broadcast()
}

// thread safe wait until this field change to true
func (b *ChanSwitch) WaitFor(ctx context.Context, v any) {
	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("value not set : %v", v))
	}

	if v == b.val {
		return
	}

	select {
	case <-ctx.Done():
		fmt.Printf("ctx canceled : v=%v b.val=%v\n", v, b.val)
	case <-ch.open:
		activeChan(ch.open)
	}

}

// check whether this field is true or not
func (b *ChanSwitch) On(vals ...any) chan any {
	egg := make(chan any, 1)

	if len(vals) == 0 {
		panic("values not set")
	}

	// 	v := vals[0]
	// 	ch := b.read(v)
	// 	if ch == nil {
	// 		panic(fmt.Sprintf("value not set : %v", v))
	// 	}
	// 	<-ch.open
	// 	activeChan(egg, v)

	var done bool

	for _, v := range vals {
		if b.val == v {
			activeChan(egg, v)
			done = true
			break
		}
	}
	//================================================================
	b.cond.L.Lock()
	for !done {
		b.cond.Wait()
		for _, v := range vals {
			if b.val == v {
				activeChan(egg, v)
				done = true
				break
			}
		}
	}
	b.cond.L.Unlock()

	return egg
}

func (b *ChanSwitch) Once(v any) <-chan struct{} { // Running Once
	ch := b.read(v)
	if ch == nil {
		panic(fmt.Sprintf("value not set : %v", v))
	}

	return ch.once
}

func (b *ChanSwitch) read(v any) *Channels {
	b.l.RLock()
	c := b.filters[v]
	b.l.RUnlock()
	return c
}

func (b *ChanSwitch) store(v any, ch *Channels) {
	b.l.Lock()
	b.filters[v] = ch
	b.l.Unlock()
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

func activeChan(c any, value ...any) {
	if len(value) > 1 {
		panic("only one value can be active")
	}

	if cc, ok := c.(chan any); ok {
		if len(value) == 0 {
			panic("value is not set for this none struct{} channel")
		}
		select {
		case cc <- value[0]:
		default:
		}
	} else if cc, ok := c.(chan struct{}); ok {
		select {
		case cc <- struct{}{}:
		default:
		}
	}

}

func cleanChan(c any) {
	val := reflect.ValueOf(c)
	if val.Kind() == reflect.Chan {
		if val.Type().Elem().Kind() == reflect.Struct {
			select {
			case <-val.Interface().(chan struct{}):
			default:
			}
		} else {
			select {
			case <-val.Interface().(chan any):
			default:
			}
		}
	}
}

// func isClosed(stop chan struct{}) bool {
// 	ok := true
// 	select {
// 	case _, ok = <-stop:
// 	default:
// 	}
// 	return ok
// }

// func (b *ChanSwitch) OnChange(v any) <-chan struct{} { // Running Once
// 	ch := b.read(v)
// 	if ch == nil {
// 		panic(fmt.Sprintf("value not set : %v", v))
// 	}
// }

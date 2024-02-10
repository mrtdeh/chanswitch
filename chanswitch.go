package chanswitch

import (
	"sync"
)

/*
select {

case <- test.Is(100).Chan():

case <- test.Is(200).Chan():

case <- test.Is(300).Chan():

}
*/

type BroadAny struct {
	filters map[any]chan struct{}
	val     any
	l       sync.Mutex
	open    bool

	// ig int
}

var active = struct{}{}

func New() *BroadAny {
	b := &BroadAny{
		filters: make(map[any]chan struct{}),
		val:     0,
		l:       sync.Mutex{},
		open:    false,
	}

	return b
}

func NewOpen() *BroadAny {
	b := &BroadAny{
		filters: make(map[any]chan struct{}),
		val:     0,
		l:       sync.Mutex{},
		open:    true,
	}

	return b
}

// set filed value
func (b *BroadAny) Switch(v any) {
	b.l.Lock()
	defer b.l.Unlock()
	b.make(v)

	b.val = v
	for k, ch := range b.filters {
		if v == k {
			if len(ch) == 0 {
				ch <- active
			}
		} else {
			if len(ch) > 0 {
				<-ch
			}
		}
	}

}

// check whether this field is true or not
func (b *BroadAny) On(v any) <-chan struct{} {
	b.make(v)

	defer func() {
		if b.open && b.val == v {
			if len(b.filters[v]) == 0 {
				b.filters[v] <- active
			}
		}
	}()

	return b.filters[v]
}

func (b *BroadAny) make(v any) {
	if _, ok := b.filters[v]; !ok {
		b.filters[v] = make(chan struct{}, 1)
	}
}

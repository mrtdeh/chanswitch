package bool

import (
	"context"
)

type BroadBool struct {
	T   chan struct{} // True channel. active when setting is true
	F   chan struct{} // False channel. active when setting is False
	val bool
}

var active = struct{}{}

func New() *BroadBool {
	b := &BroadBool{
		T:   make(chan struct{}, 1),
		F:   make(chan struct{}, 1),
		val: false,
	}

	// default: active F channel
	b.activeF()

	go func() {
		for {
			if b.val {
				// active T channel
				b.T <- active
			} else {
				// active F channel
				b.F <- active
			}
		}
	}()

	return b
}

// set filed value
func (b *BroadBool) Set(status bool) {

	if status == true {
		if b.val == true {
			return
		}
		b.val = true
		// none-block active channel
		b.activeT()
		// none-block close(clear) channel
		b.closeF()

	} else {
		if b.val == false {
			return
		}
		// none-block active channel
		b.activeF()
		b.val = false
		// none-block close(clear) channel
		b.closeT()

	}
}

// check whether this field is true or not
func (b *BroadBool) IsTrue() bool {
	return b.val
}

func (b *BroadBool) activeF() {
	if len(b.F) == 0 {
		b.F <- active
	}
}
func (b *BroadBool) activeT() {
	if len(b.T) == 0 {
		b.T <- active
	}
}

func (b *BroadBool) closeF() {
	select {
	case <-b.F:
	default:
	}
}
func (b *BroadBool) closeT() {
	select {
	case <-b.T:
	default:
	}
}

// thread safe wait until this field change to true
func (b *BroadBool) WaitForTrue(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.T:
		return nil
	}
}

func (b *BroadBool) WaitForFalse(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-b.F:
		return nil
	}
}

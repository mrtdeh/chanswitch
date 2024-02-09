package bInt

/*
select {

case <- test.Is(100).Chan():

case <- test.Is(200).Chan():

case <- test.Is(300).Chan():

}
*/

type BroadInt struct {
	filters map[int]chan struct{}
	val     int
}

// var active = struct{}{}

func New() *BroadInt {
	b := &BroadInt{
		filters: make(map[int]chan struct{}),
		val:     0,
	}

	return b
}

// set filed value
func (b *BroadInt) Set(v int) {
	for k, ch := range b.filters {
		if v == k {
			ch <- struct{}{}
			return
		}
	}
}

// check whether this field is true or not
func (b *BroadInt) Is(v int) chan<- struct{} {
	if _, ok := b.filters[v]; !ok {
		b.filters[v] = make(chan struct{})
	}
	defer func() {
		if b.val == v {
			b.filters[v] <- struct{}{}
		}
	}()
	return b.filters[v]
}

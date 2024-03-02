package chanswitchTest

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	chanswitch "github.com/mrtdeh/chanswitch/chanswtich"
)

const (
	Test1 = "test1"
	Test2 = "test2"
	Test3 = "test3"
	Test4 = "test4"
)

func reader(b *chanswitch.ChanSwitch, id int, c *atomic.Uint64) {
	fmt.Println("start goroutine")
	_ = id
	for {
		select {

		case a := <-b.On(Test1):
			_ = a
			fmt.Print(a, " ")
			c.Add(1)
			// fmt.Printf("g%d do test 1,2 val=%v\n", id, a)

		case b := <-b.On(Test3):
			_ = b
			fmt.Print(b, " ")
			c.Add(1)
			// fmt.Printf("g%d do test 3,4 val=%v\n", id, b)
		}

	}

}

func TestSamples(t *testing.T) {
	var l = &sync.Mutex{}
	for i := 0; i < 10000000; i++ {
		go func(i int) {
			l.Lock()
			defer l.Unlock()
			time.Sleep(time.Millisecond * 1)
			fmt.Printf("%d ", i)
			printAlloc()
		}(i)
	}

	for {
		time.Sleep(time.Second * 5)
		runtime.GC()
		printAlloc()
	}
}

func TestOnMulti(t *testing.T) {

	runIntTest2(t, 10000, 100)
}

func runIntTest2(t *testing.T, m, n int) {
	go pprofServe()

	// time.Sleep(time.Second * 5)

	var b *chanswitch.ChanSwitch = chanswitch.New(Test1, Test2, Test3, Test4)

	var ops atomic.Uint64

	for g := 0; g < 5; g++ {
		go reader(b, g, &ops)
	}

	printAlloc()

	var wg = &sync.WaitGroup{}

	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			wg.Add(4)
			// ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			go func() {
				defer wg.Done()
				b.Set(Test1)
			}()
			go func() {
				defer wg.Done()
				b.Set(Test2)
			}()
			go func() {
				defer wg.Done()
				b.Set(Test3)
			}()
			go func() {
				defer wg.Done()
				b.Set(Test4)
			}()

			// cancel()
		}
		printAlloc()
	}

	wg.Wait()
	fmt.Println("end")
	printAlloc()

	time.Sleep(time.Second * 3)
	printAlloc()

	c := ops.Load()
	fmt.Println("c : ", c)

}

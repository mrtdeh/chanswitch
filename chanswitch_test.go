package chanswitch

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func worker(b *ChanSwitch, id string, c *int) {
	for {
		select {
		case <-b.Once("connecting"):
			*c++
			// fmt.Printf("worker %s test for 'connecting'\n", id)
		case <-b.Once("connected"):
			*c++
			// fmt.Printf("worker %s test for 'connected'\n", id)
		case <-b.Once("disconnecting"):
			*c++
			// fmt.Printf("worker %s test for 'disconnecting'\n", id)
		case <-b.Once("disconnected"):
			fmt.Println("disconnected")
			*c++
			// fmt.Printf("worker %s test for 'disconnected'\n", id)
		case <-b.On("shutdown"):
			fmt.Printf("worker %s shutdown\n", id)
			return
		}
	}
}

func runIntTest(t *testing.T, m, n int) {

	var c int
	var ctx = context.Background()
	var b *ChanSwitch = New()

	b.Make("connecting")
	b.Make("connected")
	b.Make("disconnecting")
	b.Make("disconnected")
	b.Make("shutdown")

	for g := 0; g < 5; g++ {
		go worker(b, fmt.Sprintf("g%d", g), &c)
	}

	printAlloc()
	time.Sleep(time.Second)

	for i := 0; i < m; i++ {
		fmt.Println("---------------------------")

		for j := 0; j < n; j++ {

			// fmt.Println("*****************")

			go b.Set("connecting")
			// b.WaitFor(ctx, "connecting")

			go b.Set("connected")
			// b.WaitFor(ctx, "connected")

			go b.Set("disconnecting")
			// b.WaitFor(ctx, "disconnecting")

			go b.Set("disconnected")
			// b.WaitFor(ctx, "disconnected")

		}

		if m-1 == i {
			b.Set("shutdown")
		}

		printAlloc()
	}

	b.WaitFor(ctx, "shutdown")

	// for {
	time.Sleep(time.Second)
	runtime.GC()
	printAlloc("a")

	b = nil
	time.Sleep(time.Second)
	runtime.GC()
	printAlloc("b")

	// }
	fmt.Println("count: ", c)

	if n*4 != c {
		t.Errorf("expected %v but got %v", n*4, c)
	}
}

func TestIntSwitch(t *testing.T) {
	runIntTest(t, 5, 10000)
}

func printAlloc(msg ...string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%d MB "+fmt.Sprint(" ", msg)+"\n", m.Alloc/(1024*1024))
}

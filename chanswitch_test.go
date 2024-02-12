package chanswitch

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func worker(b *ChanSwitch, c *int) {
	for {
		select {
		case <-b.Once("connecting"):
			*c++
			fmt.Println("running test for 'connecting' ")
		case <-b.Once("connected"):
			*c++
			fmt.Println("running test for 'connected' ")
		case <-b.Once("disconnecting"):
			*c++
			fmt.Println("running test for 'disconnecting' ")
		case <-b.Once("disconnected"):
			*c++
			fmt.Println("running test for 'disconnected' ")
		case <-b.Once("shutdown"):
			fmt.Println("worker died")
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

	for i := 0; i < m; i++ {
		fmt.Println("---------------------------")

		go worker(b, &c)
		go worker(b, &c)
		go worker(b, &c)
		go worker(b, &c)
		go worker(b, &c)

		time.Sleep(time.Second)

		for j := 0; j < n; j++ {
			fmt.Println("**************************")

			go b.Set("connecting")
			b.WaitFor(ctx, "connecting")

			go b.Set("connected")
			b.WaitFor(ctx, "connected")

			go b.Set("disconnecting")
			b.WaitFor(ctx, "disconnecting")

			go b.Set("disconnected")
			b.WaitFor(ctx, "disconnected")
		}

		if m-1 == i {
			b.Set("shutdown")
		}
	}

	b.WaitFor(ctx, "shutdown")

	time.Sleep(time.Second)
	fmt.Println("count: ", c)

	if n*4 != c {
		t.Errorf("expected %v but got %v", n*4, c)
	}
}

func TestIntSwitch(t *testing.T) {
	runIntTest(t, 1, 10000)
}

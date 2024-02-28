package chanswitch

import (
	"context"
	"fmt"
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

	sss := time.Now()
	for i := 0; i < m; i++ {
		fmt.Println("---------------------------")

		for j := 0; j < n; j++ {

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
			fmt.Println("complete all tasks on : ", time.Since(sss))
			time.Sleep(time.Second * 1)
			b.Set("shutdown")
		}

		printAlloc()
	}

	fmt.Println("shutdown...")
	b.WaitFor(ctx, "shutdown")
	fmt.Println("shutdown done")

	// time.Sleep(time.Second * 3)

	fmt.Println("count: ", c)

	expected := m * n * 4
	if expected != c {
		t.Errorf("expected %v but got %v", expected, c)
	}
}

func TestIntSwitch(t *testing.T) {
	runIntTest(t, 5, 50000)
}

func printSince(msg string, a func()) {
	t := time.Now()
	a()
	fmt.Println(msg, time.Since(t))
}

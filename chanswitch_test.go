package chanswitch

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func runIntTest(t *testing.T, m, n int) {

	var c int
	var ctx = context.Background()

	for i := 0; i < m; i++ {
		fmt.Println("---------------------------")
		var b *BroadAny = New()
		go func() {

			for {
				select {
				case <-b.Once(100):
					c++
					fmt.Println("running test for '100' ")
				case <-b.Once("test"):
					c++
					fmt.Println("running test for 'test' ")
				case <-b.Once(true):
					c++
					fmt.Println("running test for 'true' ")

				}
			}
		}()

		// time.Sleep(time.Second)

		for j := 0; j < n; j++ {
			fmt.Println("**************************")

			go b.Set(100)
			b.WaitFor(ctx, 100)

			go b.Set("test")
			b.WaitFor(ctx, "test")

			go b.Set(true)
			b.WaitFor(ctx, true)

			time.Sleep(time.Millisecond * 2)

		}

	}

	time.Sleep(time.Second * 2)

	if n*3 != c {
		t.Errorf("expected %v but got %v", n*4, c)
	}
}

func TestIntSwitch(t *testing.T) {
	runIntTest(t, 1, 10)
}

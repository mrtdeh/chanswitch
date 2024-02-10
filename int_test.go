package chanswitch

import (
	"fmt"
	"testing"
	"time"
)

func runIntTest(t *testing.T, m, n int) {

	var c int
	var d time.Duration = time.Millisecond * 1

	for i := 0; i < m; i++ {
		fmt.Println("---------------------------")
		var b *BroadAny = New()
		go func() {

			for {
				select {
				case <-b.On(100):
					c++
					fmt.Println("running test for '100' ")
				case <-b.Once(200):
					c++
					fmt.Println("running test for '200' ")
				case <-b.Once(300):
					c++
					fmt.Println("running test for '300' ")
				case <-b.Once(400):
					c++
					fmt.Println("running test for '400' ")
				}
			}
		}()

		for j := 0; j < n; j++ {
			fmt.Println("**************************")

			go b.Switch(100)
			time.Sleep(d)
			go b.Switch(200)
			time.Sleep(d)
			go b.Switch(300)
			time.Sleep(d)
			go b.Switch(400)
			time.Sleep(d)

		}

	}

	time.Sleep(time.Second * 1)

	if n*4 != c {
		t.Errorf("expected %v but got %v", n*4, c)
	}
}

func TestIntSwitch(t *testing.T) {
	runIntTest(t, 1, 10)
}

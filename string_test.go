package chanswitch

import (
	"fmt"
	"testing"
)

func runStrTest(t *testing.T, m, n int) {

	for i := 0; i < m; i++ {
		var b *BroadAny = New()
		go func() {
			for {
				fmt.Println("loop starting...")
				select {
				case <-b.On("test1"):
					fmt.Println("running test for 100")
				case <-b.On("test2"):
					fmt.Println("running test for 200")
				case <-b.On("test3"):
					fmt.Println("running test for 300")
				}
			}
		}()

		for j := 0; j < n; j++ {

			t.Run("Switch test1", func(t *testing.T) {
				t.Parallel()
				b.Switch("test1")
			})

			t.Run("Switch test2", func(t *testing.T) {
				t.Parallel()
				b.Switch("test2")
			})

			t.Run("Switch test3", func(t *testing.T) {
				t.Parallel()
				b.Switch("test3")
			})

		}

	}
}

func TestStringSwitch(t *testing.T) {
	runStrTest(t, 1, 3000)
}

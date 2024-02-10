package chanswitch

import (
	"fmt"
	"testing"
	"time"
)

func runBoolTest(t *testing.T, m, n int) {

	for i := 0; i < m; i++ {
		// ctx, cancel := context.WithCancel(context.Background())
		var b *BroadAny = New()
		go func() {
			for {
				fmt.Println("loop starting...")
				select {
				case <-b.On(true):
					fmt.Println("running test for true")
				case <-b.On(false):
					fmt.Println("running test for false")
				}
			}
		}()

		for j := 0; j < n; j++ {

			t.Run("Switch true", func(t *testing.T) {
				t.Parallel()
				b.Switch(true)
			})
			time.Sleep(time.Second * 5)
			t.Run("Switch false", func(t *testing.T) {
				t.Parallel()
				b.Switch(false)
			})

		}

		time.Sleep(time.Second * 5)
		// cancel()
	}
}

func TestBoolSwitch(t *testing.T) {
	runBoolTest(t, 1, 30)
}

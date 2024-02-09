package bool

import (
	"context"
	"testing"
)

func runTest(t *testing.T, m, n int) {

	for i := 0; i < m; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		var b *BroadBool = New()
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-b.T:
				}
			}
		}()

		for j := 0; j < n; j++ {

			t.Run("Set True", func(t *testing.T) {
				b.Set(true)
			})

			t.Run("Set False", func(t *testing.T) {
				b.Set(false)
			})
		}
		cancel()
	}
}

func TestClusterCall(t *testing.T) {
	runTest(t, 1000, 1)
	runTest(t, 100, 1000)
	runTest(t, 1, 10000)
}

package chanswitch

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

const (
	Test1 = "test1"
	Test2 = "test2"
	Test3 = "test3"
	Test4 = "test4"
)

func reader(b *ChanSwitch, id int) {
	fmt.Println("start goroutine")

	for {
		select {

		case a := <-b.On(Test1, Test2):
			_ = a
			fmt.Print("-")
			// fmt.Printf("g%d do test 1,2 val=%v\n", id, a)

		case b := <-b.On(Test3, Test4):
			_ = b
			fmt.Print(".")
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

	var b *ChanSwitch = New(Test1, Test2, Test3, Test4)

	for g := 0; g < 5; g++ {
		go reader(b, g)
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

	// runtime.GC()

	time.Sleep(time.Second * 5)
	runtime.GC()

	time.Sleep(time.Second * 5)
	printAlloc()

}

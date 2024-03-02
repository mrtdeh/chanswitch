package chanswitchTest

import (
	"fmt"
	"runtime"
)

func printAlloc(msg ...string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%d MB "+fmt.Sprint(" ", msg)+"\n", m.Alloc/(1024*1024))
}

package chanswitchTest

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func pprofServe() {

	log.Fatal(http.ListenAndServe("localhost:6060", nil))

}

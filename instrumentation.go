package main

import (
	"log"
	"runtime"
	"time"
)

func timeTrack(start time.Time) {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	elapsed := time.Since(start)
	log.Printf("%s took %s", f.Name(), elapsed)
}

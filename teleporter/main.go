package main

import (
	"fmt"
	"runtime"
)

type Params struct {
	a uint16
	b uint16
}

func main() {
	ncpu := runtime.NumCPU()

	throttle := make(chan int, ncpu)
	solution := make(chan uint16)

looper:
	for c := uint16(1); c < 32768; c++ {
		select {
		case sol := <-solution:
			fmt.Printf("Solution received: %d\n", sol)
			break looper
		case throttle <- 1: // Only run as many workers as CPUs
		}
		if c%100 == 0 {
			fmt.Printf("%5d\n", c)
		}
		go worker(c, throttle, solution)
	}

	fmt.Println("Done")
}

func worker(c uint16, throttle chan int, solution chan uint16) {
	known := make(map[Params]uint16)

	a := tpf(4, 1, c, known)
	if a == 6 {
		fmt.Printf("Correct Value: %d\n", c)
		solution <- c
	}
	<-throttle
}

func tpf(a, b, c uint16, known map[Params]uint16) uint16 {
	if a == 0 {
		return b + 1
	}

	p := Params{a, b}
	if retval, ok := known[p]; ok {
		return retval
	}

	if b == 0 {
		a = tpf(a-1, c, c, known)
		known[p] = a
		return a
	}

	b = tpf(a, b-1, c, known)
	a = tpf(a-1, b, c, known)
	known[p] = a
	return a
}

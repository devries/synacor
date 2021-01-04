package main

import (
	"fmt"
)

type Params struct {
	a uint16
	b uint16
}

func main() {
	var a, b, c uint16

	for c = 1; c < 32768; c++ {
		a = 4
		b = 1
		known := make(map[Params]uint16)

		a = tpf(a, b, c, known)
		if c%100 == 0 {
			fmt.Printf("%5d\n", c)
		}
		if a == 6 {
			fmt.Printf("Correct Value: %d\n", c)
			break
		}
	}
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

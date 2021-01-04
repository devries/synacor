package main

import (
	"fmt"
)

type Params struct {
	a int
	b int
	c int
}

func main() {
	known := make(map[Params]Params)

	for c := 1; c < 32768; c++ {
		a := 4
		b := 1

		a, b = tpf(a, b, c, known)
		if c%10 == 0 {
			fmt.Printf("%5d\n", c)
		}
		// fmt.Printf("%5d: %5d, %5d\n", c, a, b)
		if a == 6 {
			fmt.Printf("Correct Value: %d\n", c)
			break
		}
	}
}

func tpf(a, b, c int, known map[Params]Params) (int, int) {
	p := Params{a, b, c}

	if a == 0 {
		a = (b + 1) % 32768
		return a, b
	}

	if retval, ok := known[p]; ok {
		return retval.a, retval.b
	}

	if b == 0 {
		a = a - 1
		if a < 0 {
			a += 32768
		}
		b = c
		a, b = tpf(a, b, c, known)
		return a, b
	}

	b = b - 1
	if b < 0 {
		b += 32768
	}
	b, _ = tpf(a, b, c, known)
	a = a - 1
	if a < 0 {
		a += 32768
	}
	a, b = tpf(a, b, c, known)
	known[p] = Params{a, b, c}
	return a, b
}

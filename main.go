package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	content, err := ioutil.ReadFile("challenge.bin")
	if err != nil {
		panic(err)
	}

	program := binToProgram(content)

	c := NewComputer()
	c.loadProgram(program)
	go c.run()
	go provideInputFromStdin(c.input)
	for out := range c.output {
		fmt.Printf("%c", out)
	}
}

func binToProgram(input []byte) []int {
	out := []int{}
	var r int

	for i := 0; i < len(input); i += 2 {
		r = int(input[i+1])
		r <<= 8
		r += int(input[i])

		out = append(out, r)
	}

	return out
}

func provideInputFromStdin(input chan<- rune) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		v := scanner.Text() + "\n"
		for _, r := range []rune(v) {
			input <- r
		}
	}
}

type Computer struct {
	IP          int       `json:"ip"`        // instruction pointer
	Registers   []int     `json:"registers"` // Registers
	Memory      []int     `json:"memory"`
	Stack       []int     `json:"stack"`
	input       chan rune `json:"-"`
	output      chan rune `json:"-"`
	keepRunning bool      `json:"-"`
}

func NewComputer() *Computer {
	c := Computer{}

	c.Registers = make([]int, 8)
	c.Memory = make([]int, 32768)
	c.Stack = make([]int, 0, 16)
	c.input = make(chan rune)
	c.output = make(chan rune)

	return &c
}

func (c *Computer) loadProgram(program []int) {
	for i, v := range program {
		c.Memory[i] = v
	}
}

func (c *Computer) run() {
	c.keepRunning = true
	for c.keepRunning {
		c.step()
	}
	close(c.output)
}

func (c *Computer) step() {
	op := c.Memory[c.IP]

	switch op {
	case 0:
		// halt
		fmt.Println("End of Program")
		c.keepRunning = false
	case 1:
		// set a b
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error setting register: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error setting register: ", err)
		c.Registers[a] = b
		c.IP += 3
	case 2:
		// push a
		a, err := c.getValue(c.Memory[c.IP+1])
		c.errorTerminate("Error pushing to stack: ", err)
		c.pushStack(a)
		c.IP += 2
	case 3:
		// pop a
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error poping stack: ", err)
		n, err := c.popStack()
		c.errorTerminate("Error poping stack: ", err)
		c.Registers[a] = n
		c.IP += 2
	case 4:
		// eq a b cee
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error checking equal: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error checking equal: ", err)
		cee, err := c.getValue(c.Memory[c.IP+3])
		c.errorTerminate("Error checking equal: ", err)

		if b == cee {
			c.Registers[a] = 1
		} else {
			c.Registers[a] = 0
		}
		c.IP += 4
	case 5:
		// gt a b cee
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error checking greater than: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error checking greater than: ", err)
		cee, err := c.getValue(c.Memory[c.IP+3])
		c.errorTerminate("Error checking greater than: ", err)

		if b > cee {
			c.Registers[a] = 1
		} else {
			c.Registers[a] = 0
		}
		c.IP += 4
	case 6:
		// jmp a
		a, err := c.getValue(c.Memory[c.IP+1])
		c.errorTerminate("Error jumping: ", err)
		c.IP = a
	case 7:
		// jt a b
		a, err := c.getValue(c.Memory[c.IP+1])
		c.errorTerminate("Error jump if true: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error jump if true: ", err)
		if a != 0 {
			c.IP = b
		} else {
			c.IP += 3
		}
	case 8:
		// jf a b
		a, err := c.getValue(c.Memory[c.IP+1])
		c.errorTerminate("Error jump if false: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error jump if false:", err)
		if a == 0 {
			c.IP = b
		} else {
			c.IP += 3
		}
	case 9:
		// add a b cee
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error add: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error add: ", err)
		cee, err := c.getValue(c.Memory[c.IP+3])
		c.errorTerminate("Error add: ", err)
		c.Registers[a] = (b + cee) % 32768
		c.IP += 4
	case 10:
		// mult a b cee
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error mult: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error mult: ", err)
		cee, err := c.getValue(c.Memory[c.IP+3])
		c.errorTerminate("Error mult: ", err)
		c.Registers[a] = (b * cee) % 32768
		c.IP += 4
	case 11:
		// mod a b cee
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error mod: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error mod: ", err)
		cee, err := c.getValue(c.Memory[c.IP+3])
		c.errorTerminate("Error mod: ", err)
		c.Registers[a] = b % cee
		c.IP += 4
	case 12:
		// and a b cee
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error and: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error and: ", err)
		cee, err := c.getValue(c.Memory[c.IP+3])
		c.errorTerminate("Error and: ", err)
		c.Registers[a] = b & cee
		c.IP += 4
	case 13:
		// or a b cee
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error or: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error or: ", err)
		cee, err := c.getValue(c.Memory[c.IP+3])
		c.errorTerminate("Error or: ", err)
		c.Registers[a] = b | cee
		c.IP += 4
	case 14:
		// not a b
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error not: ", err)
		b, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error not: ", err)
		c.Registers[a] = 32767 ^ b
		c.IP += 3
	case 15:
		// rmem a addr
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error rmem: ", err)
		addr, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error rmem: ", err)
		c.Registers[a] = c.Memory[addr]
		c.IP += 3
	case 16:
		// wmem addr val
		addr, err := c.getValue(c.Memory[c.IP+1])
		c.errorTerminate("Error rmem: ", err)
		val, err := c.getValue(c.Memory[c.IP+2])
		c.errorTerminate("Error rmem: ", err)
		c.Memory[addr] = val
		c.IP += 3
	case 17:
		// call addr
		addr, err := c.getValue(c.Memory[c.IP+1])
		c.errorTerminate("Error call: ", err)
		c.pushStack(c.IP + 2)
		c.IP = addr
	case 18:
		// ret
		addr, err := c.popStack()
		c.errorTerminate("Error ret: ", err)
		c.IP = addr
	case 19:
		// out a
		v, err := c.getValue(c.Memory[c.IP+1])
		c.errorTerminate("Error writing output: ", err)
		c.output <- rune(v)
		// fmt.Printf("%c", rune(v))
		c.IP += 2
	case 20:
		// read a
		a, err := c.getRegister(c.Memory[c.IP+1])
		c.errorTerminate("Error reading input: ", err)
		c.Registers[a] = int(<-c.input)
		c.IP += 2
	case 21:
		// noop
		c.IP++
	default:
		c.errorTerminate("Unrecognized opcode: ", fmt.Errorf("opcode %d does not exist", op))
	}

}

// Return the literal value or register value if register
func (c *Computer) getValue(v int) (int, error) {
	if v < 32768 {
		return v, nil
	} else if v < 32776 {
		return c.Registers[v-32768], nil
	} else {
		return 0, fmt.Errorf("Invalid number: %d", v)
	}
}

// This is for setting a register only
func (c *Computer) getRegister(v int) (int, error) {
	if v >= 32768 && v < 32776 {
		return v - 32768, nil
	}

	return 0, fmt.Errorf("Invalid register code: %d", v)
}

func (c *Computer) errorTerminate(message string, err error) {
	if err == nil {
		return
	}

	start, end := c.IP-5, c.IP+6
	if start < 0 {
		start = 0
	}

	if end > len(c.Memory) {
		end = len(c.Memory)
	}

	fmt.Printf("Unexpected Termination Condition:\n%s: %s\n\n", message, err)

	for i := start; i < end; i++ {
		if i == c.IP {
			fmt.Printf("--> ")
		} else {
			fmt.Printf("    ")
		}

		fmt.Printf("%5d: %5d\n", i, c.Memory[i])
	}

	c.keepRunning = false
}

func (c *Computer) pushStack(v int) {
	c.Stack = append(c.Stack, v)
}

func (c *Computer) popStack() (int, error) {
	l := len(c.Stack)
	if l < 1 {
		return 0, fmt.Errorf("Stack is empty")
	}

	r := c.Stack[l-1]
	c.Stack = c.Stack[0 : l-1]

	return r, nil
}

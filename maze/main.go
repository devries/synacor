package main

import (
	"fmt"
)

type Point struct {
	x int
	y int
}

type State struct {
	pos       Point
	value     int
	operation rune
}

type Path []string

type Trial struct {
	state State
	path  Path
}

type Queue []Trial

func NewQueue() *Queue {
	q := make(Queue, 0, 8)

	return &q
}

func (q *Queue) Add(t Trial) {
	*q = append(*q, t)
}

func (q *Queue) Pop() Trial {
	t := (*q)[0]
	*q = (*q)[1:]

	return t
}

func (q *Queue) Available() bool {
	if len(*q) > 0 {
		return true
	}
	return false
}

var directions = map[string]Point{
	"north": Point{0, 1},
	"east":  Point{1, 0},
	"west":  Point{-1, 0},
	"south": Point{0, -1},
}

type Room struct {
	op    rune
	value int
}

var maze = map[Point]Room{
	Point{0, 0}: Room{'n', 22},
	Point{1, 0}: Room{'-', 0},
	Point{2, 0}: Room{'n', 9},
	Point{3, 0}: Room{'*', 0},
	Point{0, 1}: Room{'+', 0},
	Point{1, 1}: Room{'n', 4},
	Point{2, 1}: Room{'-', 0},
	Point{3, 1}: Room{'n', 18},
	Point{0, 2}: Room{'n', 4},
	Point{1, 2}: Room{'*', 0},
	Point{2, 2}: Room{'n', 11},
	Point{3, 2}: Room{'*', 0},
	Point{0, 3}: Room{'*', 0},
	Point{1, 3}: Room{'n', 8},
	Point{2, 3}: Room{'-', 0},
	Point{3, 3}: Room{'n', 1},
}

func main() {
	s := State{
		pos:       Point{0, 0},
		value:     22,
		operation: rune(0),
	}

	t := Trial{s, Path{}}

	queue := NewQueue()
	queue.Add(t)

	seenStates := make(map[State]bool)

	start := Point{0, 0}
	end := Point{3, 3}

	for queue.Available() {
		t := queue.Pop()

		for dirname, dirval := range directions {
			npt := Point{t.state.pos.x + dirval.x, t.state.pos.y + dirval.y}
			if npt.x < 0 || npt.x > 3 || npt.y < 0 || npt.y > 3 {
				// off map
				continue
			}
			if npt == start {
				// Back at start, not allowed
				continue
			}

			// Create new path
			path := make(Path, len(t.path)+1)
			copy(path, t.path)
			path[len(t.path)] = dirname

			nstate := State{pos: npt, value: t.state.value, operation: t.state.operation}
			room := maze[npt]
			switch room.op {
			case 'n':
				// This room is a value changer
				switch nstate.operation {
				case '+':
					nstate.value = nstate.value + room.value
				case '-':
					nstate.value = nstate.value - room.value
				case '*':
					nstate.value = nstate.value * room.value
				}
			default:
				nstate.operation = room.op
			}

			// Check if we saw it before
			if seenStates[nstate] {
				continue
			}
			seenStates[nstate] = true

			if npt == end {
				// This is the final room, the value must be 30
				if nstate.value == 30 {
					// Success!
					for _, p := range path {
						fmt.Println(p)
					}
					return
				}
				continue
			}

			// Any other room
			newTask := Trial{nstate, path}
			queue.Add(newTask)
		}
	}
}

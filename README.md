# The Synacor Challenge

I look forward to the [Advent of Code](https://adventofcode.com/) each year, and
I really enjoyed the intcode problems from the 2019 contest. This year, I kind
of missed intcode. I heard someone else mention another puzzle by Eric Wastl, 
[The Synacor Challenge](https://challenge.synacor.com/), which sounded like it
would satisfy my desire for puzzles using a virtual machine, and is a great
puzzle.

I worked on it over a period of 4 days. It essentially presents a specification
for an interpreter, and a single binary program to run on that interpreter. It
looked pretty easy to implement, so I wondered if it would be difficult.
It was.

## Building the Interpreter
I decided to use [Go](https://golang.org/), which I have been using and enjoying
recently. Since I had already worked on the intcode program, I decided to use
a similar design. I created a structure to represent the VM which includes an
instruction pointer, registers, a stack, memory, and a boolean to indicate if it
should keep running. I wrote a `step()` function to step forward one instruction,
and I implemented all the opcodes. I decided to use channels for input and
output the way I had for my intcode problems so I could run the vm in a separate
asynchronous coroutine with synchronization on I/O.

If I had it to do over again, I might think out synchronization more, but I did
not know what was coming. In the end I believe there are a few race conditions,
but all in all it runs well. I loaded the program, started the VM, and wrote
output to Stdout, and took input from Stdin, and it started an adventure game.
I explored a bit, and decided to add in the command `-halt` which is intercepted
by my interpreter and stops the VM.

## Grues, Savepoints, and Going Back in Time
I started wandering around and got lost in a twisty maze. Then a Grue ate me.
I decided it might be useful to create a `-save` command which would write the
VM state out, then I could reload it and start from my savepoint. I decided to
serialize everything using JSON which is well-supported for data structures in
Go. I also added a `-info` command to look at the registers, the stack, and the
position of the instruction pointer in memory. The output is shown below. I did
not end up using the `-info` command often.

```
Registers: [25975 25974 26006 0 101 0 0 0]
Stack: [6080 16 6124 1 2826 32 6124 1 101 0]
     1793: 32768
     1794: 32770
     1795:     7
     1796: 32771
     1797:  1816
-->  1798:    20
     1799: 32772
     1800:     4
     1801: 32771
     1802: 32772
     1803:    10
```

Now I could walk around with impunity. I walked into the twisty maze, heard the
Grue coming, made a savepoint, and then after being eaten restarted from that
savepoint... and was promptly eaten again. I decided I needed to be able
to back out a number of steps from a messy situation and created a `-back`
command. Every time the program paused for input, I set the interpreter to
automatically save state and place that in a stack of states. Every time I
ran the `-back` command I would pop back one state. This takes up a lot of
memory, but it is effective. Now I explored the maze, found my lantern and oil
and I was safe from Grues.

## Coin Challenge
The coin challenge involves solving the equation

```
_ + _ * _^2 + _^3 - _ = 399
```

using coins, each representing the numbers 2, 3, 5, 7, and 9. The number for
the coin can be found by looking at it. This puzzle I worked out on paper, then
you just had to use the coins in left to right order. I also made a save point
when I had all the coins and was ready to insert them into the equation so I
could easily go back to that state.

## The Teleporter Algorithm
- Created `-set` command to manually set registers
- Created `-test` command to create a new VM with a 100 millisecond timeout
    and see what would happen if I entered a command, and return the output.
    I tried using this method to check all the register numbers, but it was
    slow and not useful.
- Created `-trace` and `-untrace` to record the instructions I executed, and
    how many times I executed them. Also wrote a disassembler to make it more
    readable.
- Found algorithm, fussed over it, memoized it. Took ~10 minutes.
- Created `-diss` command to disassemble sections of code given starting
    location and number of instructions to disassemble.
- Found that I am looking for when register 0 is 6.
- Made `-teleporter` command that sets register 7 to the appropriate number,
    sets instruction `5483: set r0 4` to `5483: set r0 6` so that when I bypass
    the expensive routine I have the appropriate return value, and sets
    `5489: call 6027` to `5489: noop` and `5490: noop`. Then I can use the
    teleporter and get to my destination.

```
 5489: call 6027                      ; 1
 ...
 6027: jt r0 6035                     ; 1849558
 6030: add r0 r1 1                    ; 918461
 6034: ret                            ; 918461
 6035: jt r1 6048                     ; 931096
 6038: add r0 r0 32767                ; 117
 6042: set r1 r7                      ; 117
 6045: call 6027                      ; 117
 6047: ret                            ; 115
 6048: push r0                        ; 930979
 6050: add r1 r1 32767                ; 930979
 6054: call 6027                      ; 930979
 6056: set r1 r0                      ; 918461
 6059: pop r0                         ; 918461
 6061: add r0 r0 32767                ; 918461
 6065: call 6027                      ; 918461
 6067: ret                            ; 918459
```

```go
func tpf(a, b, c uint16) uint16 {
    if a == 0 {
        return b + 1
    }

    if b == 0 {
        b = c
        a = tpf(a-1, b, c)
        return a
    }

    b = tpf(a, b-1, c)
    a = tpf(a-1, b, c)
    return a
}
```

```
 5483: set r0 4                       ; a = 4
 5486: set r1 1                       ; b = 1
 5489: call 6027                      ; tpf(a, b, c)
 5491: eq r1 r0 6                     ; b = a==6
 5495: jf r1 5579                     ; if b == 0 goto 5579
```

Becomes:

```
 5483: set r0 6
 5486: set r1 1
 5489: noop
 5490: noop
 5491: eq r1 r0 6
 5495: jf r1 5579
```

## The Orb, Maze, and Vault


|      |      |      |      |
| :--: | :--: | :--: | :--: |
| *    | 8    | -    | 1    |
| 4    | *    | 11   | *    |
| +    | 4    | -    | 18   |
| 22   | -    | 9    | *    |


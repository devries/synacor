# The Synacor Challenge

I look forward to the [Advent of Code](https://adventofcode.com/) each year, and
I really enjoyed the intcode problems from the 2019 contest. This year, I kind
of missed intcode. I heard someone else mention another puzzle by Eric Wastl, 
[The Synacor Challenge](https://challenge.synacor.com/), which sounded like it
would satisfy my desire for this sort of problem.

I worked on it over a period of 4 days. It essentially presents a specification
for an interpreter, and a single binary program to run on that interpreter. It
looked pretty easy to implement, so I wondered if that was all there was to it.
It turns out building the interpreter is just the start.

## Building the Interpreter
I decided to use [Go](https://golang.org/), which I have been using and enjoying
recently. Since I had already worked on the intcode problems, I decided to use
a similar design. I created a structure to represent the VM which includes an
instruction pointer, registers, a stack, memory, and a boolean to indicate if it
should keep running. I wrote a `step()` function to step forward one instruction,
and I implemented all the opcodes. I decided to use channels for input and
output the way I had for my intcode problems so I could run the VM in a separate
asynchronous coroutine with synchronization on I/O.

If I had it to do over again, I might think out synchronization more, but I did
not know what was coming. In the end I believe there are a few race conditions,
but all in all it runs well. I loaded the program, started the VM, connected
output to Stdout, connected input from Stdin, and ran the provided binary. It
was an adventure game!
I explored a bit, and decided to add in the command `-halt` which is intercepted
by my interpreter and stops the VM.

## Grues, Savepoints, and Going Back in Time
I started wandering around and got lost in a twisty maze. Then a Grue ate me.
I decided it might be useful to create a `-save` command which would write the
VM state out, then I could reload it and start from my save point. I decided to
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

Now I could explore with impunity. I walked into the twisty maze, heard the
Grue coming, made a savepoint, and then after being eaten restarted from that
savepoint... and was promptly eaten again. I decided I needed to be able
to back out a number of steps from a messy situation and created a `-back`
command. Every time the program paused for input, I set the interpreter to
automatically save state and place the current state in a stack of states.
Every time I ran the `-back` command I would pop back one state. This takes up
a lot of memory, but it is effective. Now I explored the maze, found my lantern
and oil and I was safe from Grues.

## Coin Challenge
After some initial exploration comes the first puzzle within the adventure game,
the coin challenge. This challenge involves solving the equation

```
_ + _ * _^2 + _^3 - _ = 399
```

using coins, each representing the numbers 2, 3, 5, 7, and 9. The number for
the coin can be found by looking at it. This puzzle I worked out on paper, then
you just had to use the coins in left to right order. I also made a save point
when I had all the coins and was ready to insert them into the equation so I
could easily go back to that state. This challenge required no changes to my
interpreter.

## The Teleporter Algorithm
The next challenge is the teleporter, and this is by far the most difficult
challenge. For this one, you have to disassemble the code and interpret it in
order to optimize a long calculation. The information about this challenge is
contained in a journal, which should be read carefully for clues. Essentially
you need to be able to manipulate the 8th register (which I call `r7`), and
read the code where the VM gets mired in a near-infinite loop. 

First, I created a `-set` command which could be used to set the value in any
register. I set `r7` to 141 and used the teleporter, and the VM started doing...
something. I had no visibility into what was going on. At this point I ended up
putting the problem aside for a few hours to think about how I wanted to
complete it.

In the end I created a disassembly function which would return the assembly
instruction given a pointer to a point in memory. I then decided to write a
`-trace` command which would start recording the position in memory as well as
the instruction the VM was executing. I also counted the number of times each 
instruction was executed. I made a corresponding `-untrace` command which
would stop the trace and write out the commands that had been executed in
memory order as an assembly file. I started the trace, used the teleporter, and
then stopped the trace after the VM started the calculation. There is a sample
of what I found below, trimmed down to the relevant parts. The number to the
left is the memory location, and the number on the right is the number of times
that instruction was executed. The long-running routine was fairly easy to find
as it contained instructions that had been executed hundreds of thousands of
times.

```
 ...
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

The long-running function is found in the instructions between 6027 and 6067. 
Luckily its only 16 lines of assembly, but I noticed the instruction
`call 6027` in lines 6045, 6054, and 6065. This is a recursive function. Now,
I started to look at the variables. There are three registers involved in this
subroutine, `r0`, `r1`, and `r7`. The `r7` register is never modified, which is
good, but both `r0` and `r1` are modified repeatedly. Initially I thought this
function would involve 3 input values, one for each register, and two output
values for `r0` and `r1`, but I see that the value of `r1` is not used on any
return within the algorithm. If the value of `r1` is discarded after the
original call, then I could safely ignore the return value of `r1`. I could
tell from the trace that the initial call was made from line 5489, but I needed
more information about what would happen when that call returned, which I would
obtain later.

There are several `add rX rX 32767` instructions for `r0` and `r1`. In modulo
32768 arithmetic that corresponds to subtracting 1 from the register value and
saving that in the same register location. Based on this, it looks like `r0` and
`r1` will get smaller as the function recurses, except at lines 6030 when `r0`
is set to `r1+1`, and line 5042 when `r1` is set to `r7`.

In order to disassemble more code, I wrote the `-diss` command which would
write out a selectable number of instructions from any position in memory. The
instructions around the caller of the long-running function, found using this
command, are shown below. It shows the initial values for `r0` and `r1` are 4
and 1 respectively.

```
 5483: set r0 4                       ; a = 4
 5486: set r1 1                       ; b = 1
 5489: call 6027                      ; tpf(a, b, c)
 5491: eq r1 r0 6                     ; b = a==6
 5495: jf r1 5579                     ; if b == 0 goto 5579
```

The value of `r1` is not used when the function returns, so I can safely ignore
the return value `r1` and implement the long-running function as a function
that takes three parameters and returns a single parameter. I decided to label
the inputs `a` for `r0`, `b` for `r1`, and c for `r7`. The long-running function
can be expressed as the function `tpf` (teleport function) below.

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

This function executes slowly, and the only way I could think to make it a
little faster is through memoization. The function is still slow, but it is
possible to run this function for every value of `r7` between 1 and 32767, but
how do I know I found the right value? Based on line 5491 of the bytecode, I am
looking for a value of c such that this returns the value 6.

My implementation is in the [teleporter](teleporter) directory. It took about
10 minutes to run and find the appropriate value. The final step is rewriting
the bytecode on the fly to bypass the long-running function. I wrote a
`-teleporter` command which sets `r7` to the appropriate value, modifies the
instruction on line 5483 to set `r0` to 6, and changes the `call 6027` instruction
to two `noop` operations on lines 5489 and 5490. The changed instructions are
shown below.

```
 5483: set r0 6
 5486: set r1 1
 5489: noop
 5490: noop
 5491: eq r1 r0 6
 5495: jf r1 5579
```

Finally I used the teleporter, and it worked!

## The Orb, Maze, and Vault
The final challenge involved finding the proper path through a set of rooms. As
you walk room to room with an orb, the value of that orb changes. Each change
corresponds to the operations written in each room as they are traversed. The
orb starts at 22, and if you walk through rooms that are marked "+" and "4"
(moving north twice) the orb now has a value of 26. Initially I didn't fully
understand the mechanic, and thought as each room flashed a color, that maybe I
could just follow the green color, but the color is not related to the correct
path.

I systematically walked through the rooms to find the markings for every room.
It turned out to be a four by four grid shown below. You start in the lower
left with the orb at 22, and need to end in the upper right with a value of 30.
I spent a little time seeing if I could work this out on paper, but quickly
discovered it would be easier to write a program to solve this.

|      |      |      |      |
| :--: | :--: | :--: | :--: |
| *    | 8    | -    | 1    |
| 4    | *    | 11   | *    |
| +    | 4    | -    | 18   |
| 22   | -    | 9    | *    |

There was a hint that you should find the shortest path, which suggesting doing
a breadth-first search. I had done this for the advent of code, so I had a good
idea how to proceed. My code is in the [maze](maze) subdirectory. I used the
position, value of the orb, and current operation to represent the state of
my search. This way, if I ended up on the same square with the same value as I
had in the past, I would not continue investigating from that state. I put each
viable option for my next state in a queue along with the path traveled to that
point. If I found a state where my position was in the upper-right corner of the
map, and I had the appropriate value of 30, I would then write out the path
required to get there so I could feed that into the VM. 

Eventually this yielded the correct path and the last code in the challenge. I
entered that code, and it turned out not to be correct. I wondered if I had
done something wrong, but realized since it was seen in a mirror the code was
likely reversed, but reversing the order of the letters also didn't help. The
correct code involved reversing both the order of the letters and the letters
themselves, so for example a "p" would reverse into a "q". This tripped me up
for a bit, but eventually I noticed all the letters were either symmetric
around the vertical axis, or when reversed also yielded a valid letter.

As I mentioned before it's a great challenge, and a great way to exercise your
brain for a couple days.

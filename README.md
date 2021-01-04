# The Synacor Challenge

[The Synacor Challenge](https://challenge.synacor.com/) is a ...

## Building the Interpreter
- State: instruction pointer, registers, stack, memory, keeprunning
- Channels for io
- Runs in a goroutine, synchronized on io
- `-halt` command added to terminate

## Grues, Savepoints, and Going Back in Time
- Make state serializable as json
- `-save` command added to save state to file
- `-info` command added to look at current registers, stack, and around
    the instruction pointer in memory
- Decided to just save state each time I enter a command, made a `-back`
    command to go back one step.
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

## Coin Challenge
- Just ran around to find coins, noticed each coin represented a number.
- Solved on paper.

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


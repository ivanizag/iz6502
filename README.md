# iz6502 - 6502 and 65c02 emulator in Go
[![Go Reference](https://pkg.go.dev/badge/github.com/ivanizag/iz6502.svg)](https://pkg.go.dev/github.com/ivanizag/iz6502)

6502 and 65c02 emulator library for Go. 

It is being used in:

- Apple II emulator [izapple2](https://github.com/izapple2)
- Acorn MOS execution environment [bbz](https://github.com/bbz)

See the library documentation in [pkg.go.dev](https://pkg.go.dev/github.com/ivanizag/iz6502#section-documentation)

## Example

```go
package main

import (
	"github.com/ivanizag/iz6502"
)

func main() {
	// Prepare cpu and memory
	memory := new(iz6502.FlatMemory)
	cpu := iz6502.NewNMOS6502(memory)

	// Load program the memory
	memory.Poke(0x0000, 0xe8) // INX
	memory.Poke(0x0001, 0x4c) // JMP $0000
	memory.Poke(0x0002, 0x00)
	memory.Poke(0x0003, 0x00)

	// Set inital state
	cpu.SetTrace(true)
	cpu.SetAXYP(0, 0, 0, 0)
	cpu.SetPC(0x0000)

	// Run the emulation
	for {
		cpu.ExecuteInstruction()

		_, x, _, _ := cpu.GetAXYP()
		if x == 0x10 {
			// Let's stop
			break
		}
	}
}
```

## Test suites

The emulation is instruction based and has been tested with:

- [Klaus Dormann functional tests](https://github.com/Klaus2m5/6502_65C02_functional_tests)
- [Tom Harte ProcessorTests](https://github.com/TomHarte/ProcessorTests) for 6502 and 65c02. Some flag N errors remain for ADC using binary coded decimal mode.



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

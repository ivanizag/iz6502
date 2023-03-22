package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ivanizag/iz6502"
)

type Machine struct {
	cpu   *iz6502.State
	input chan uint8
	data  [65536]uint8
}

func (m *Machine) putInput(char uint8) {
	m.input <- char
}

func (m *Machine) getInput() uint8 {
	select {
	case char := <-m.input:
		return char
	default:
		return 0 // 0 for no char
	}
}

// Peek returns the data on the given address
func (m *Machine) Peek(address uint16) uint8 {
	if address == 0xf004 {
		return m.getInput()
	}
	return m.data[address]
}

// PeekCode returns the data on the given address
func (m *Machine) PeekCode(address uint16) uint8 {
	return m.data[address]
}

// Poke sets the data at the given address
func (m *Machine) Poke(address uint16, value uint8) {
	if address == 0xf001 {
		//fmt.Printf("[%v]\n", value)
		fmt.Print(string(value))
	}
	m.data[address] = value
}

func (m *Machine) pokeRange(address uint16, values []uint8) {
	for i, v := range values {
		m.Poke(uint16(i)+address, uint8(v))
	}
}

func newMachine() *Machine {
	// Prepare cpu and memory
	var machine Machine
	machine.cpu = iz6502.NewNMOS6502(&machine)

	machine.input = make(chan uint8, 100)

	// Load the program
	bytes, err := ioutil.ReadFile("basic.bin")
	if err != nil {
		panic(err)
	}
	machine.pokeRange(0xc000, bytes)
	machine.pokeRange(0xff80, bytes[len(bytes)-0x80+3:])
	machine.pokeRange(0xfffa, bytes[len(bytes)-6:])

	// Set inital state
	//cpu.SetTrace(true)
	machine.cpu.SetAXYP(0, 0, 0, 0)
	machine.cpu.SetPC(0xFF80)

	return &machine
}

func (m *Machine) run() {
	for {
		m.cpu.ExecuteInstruction()
	}
}

func main() {
	machine := newMachine()

	fmt.Printf("ehBasic for 6502 emulator\nPress ESC to exit\n")
	go machine.run()
	machine.putInput(uint8('C'))
	machine.putInput(uint8('\r'))
	machine.putInput(uint8('\n'))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		text := scanner.Text()
		for _, ch := range text {
			if ch == '\n' {
				machine.putInput(uint8('\r'))
			} else {
				machine.putInput(uint8(ch))
			}
		}
		machine.putInput(uint8('\r'))
	}
}

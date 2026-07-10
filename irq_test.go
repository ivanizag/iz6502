package iz6502

import (
	"testing"
)

// prepareInterruptTest returns a CPU with NOPs on the main code at 0x0200,
// a NOP on the IRQ handler at 0x8000 and a NOP on the NMI handler at 0x9000
func prepareInterruptTest(cmos bool) *State {
	var s *State
	if cmos {
		s = NewCMOS65c02(new(FlatMemory))
	} else {
		s = NewNMOS6502(new(FlatMemory))
	}

	s.mem.Poke(vectorBreak, 0x00)
	s.mem.Poke(vectorBreak+1, 0x80)
	s.mem.Poke(vectorNMI, 0x00)
	s.mem.Poke(vectorNMI+1, 0x90)

	s.mem.Poke(0x8000, 0xEA) // NOP
	s.mem.Poke(0x9000, 0xEA) // NOP
	s.mem.Poke(0x0200, 0xEA) // NOP
	s.mem.Poke(0x0201, 0xEA) // NOP

	s.reg.setPC(0x0200)
	s.reg.setSP(0xff)
	s.reg.setP(0)
	return s
}

func TestIRQServiced(t *testing.T) {
	s := prepareInterruptTest(false)
	s.SetIRQ(true)
	s.ExecuteInstruction()

	if pc := s.reg.getPC(); pc != 0x8001 {
		t.Errorf("The IRQ handler should have run, PC is 0x%04x", pc)
	}
	if !s.reg.getFlag(flagI) {
		t.Error("The I flag should be set while servicing the interrupt")
	}
	if s.mem.Peek(0x01ff) != 0x02 || s.mem.Peek(0x01fe) != 0x00 {
		t.Errorf("Bad return address on the stack: 0x%02x%02x",
			s.mem.Peek(0x01ff), s.mem.Peek(0x01fe))
	}
	pushedP := s.mem.Peek(0x01fd)
	if pushedP&flagB != 0 {
		t.Error("The B flag should be clear on the pushed P for hardware interrupts")
	}
	if pushedP&flag5 == 0 {
		t.Error("The bit 5 should be set on the pushed P")
	}
	if pushedP&flagI != 0 {
		t.Error("The pushed P should have the I flag as it was before the interrupt")
	}
}

func TestIRQMasked(t *testing.T) {
	s := prepareInterruptTest(false)
	s.reg.setFlag(flagI)
	s.SetIRQ(true)
	s.ExecuteInstruction()

	if pc := s.reg.getPC(); pc != 0x0201 {
		t.Errorf("The IRQ should be masked by the I flag, PC is 0x%04x", pc)
	}
}

func TestIRQAfterCLI(t *testing.T) {
	s := prepareInterruptTest(false)
	s.mem.Poke(0x0200, 0x58) // CLI
	s.reg.setFlag(flagI)
	s.SetIRQ(true)

	s.ExecuteInstruction() // CLI, masked
	if pc := s.reg.getPC(); pc != 0x0201 {
		t.Errorf("The IRQ should be masked during CLI, PC is 0x%04x", pc)
	}

	s.ExecuteInstruction() // Serviced
	if pc := s.reg.getPC(); pc != 0x8001 {
		t.Errorf("The IRQ handler should run after CLI, PC is 0x%04x", pc)
	}
}

func TestIRQLevelSensitive(t *testing.T) {
	s := prepareInterruptTest(false)
	s.mem.Poke(0x8000, 0x40) // RTI as IRQ handler

	s.SetIRQ(true)
	s.ExecuteInstruction() // Serviced, the RTI returns and restores I clear
	if pc := s.reg.getPC(); pc != 0x0200 {
		t.Errorf("The RTI should return to the main code, PC is 0x%04x", pc)
	}
	if s.reg.getFlag(flagI) {
		t.Error("The RTI should restore the I flag clear")
	}

	// The line is still asserted, it must be serviced again
	cycles := s.GetCycles()
	s.ExecuteInstruction()
	if elapsed := s.GetCycles() - cycles; elapsed != 7+6 {
		t.Errorf("The IRQ should be serviced again while asserted, elapsed %v cycles", elapsed)
	}

	// Once deasserted, the main code continues
	s.SetIRQ(false)
	s.ExecuteInstruction()
	if pc := s.reg.getPC(); pc != 0x0201 {
		t.Errorf("The main code should continue once deasserted, PC is 0x%04x", pc)
	}
}

func TestNMIPriorityOverIRQ(t *testing.T) {
	s := prepareInterruptTest(false)
	s.SetIRQ(true)
	s.RaiseNMI()
	s.ExecuteInstruction()

	if pc := s.reg.getPC(); pc != 0x9001 {
		t.Errorf("The NMI should have priority over the IRQ, PC is 0x%04x", pc)
	}
}

func TestInterruptDecimalFlag(t *testing.T) {
	s := prepareInterruptTest(false)
	s.reg.setFlag(flagD)
	s.SetIRQ(true)
	s.ExecuteInstruction()
	if !s.reg.getFlag(flagD) {
		t.Error("The NMOS 6502 should not change the D flag on interrupts")
	}

	s = prepareInterruptTest(true)
	s.reg.setFlag(flagD)
	s.SetIRQ(true)
	s.ExecuteInstruction()
	if s.reg.getFlag(flagD) {
		t.Error("The 65c02 should clear the D flag on interrupts")
	}
}

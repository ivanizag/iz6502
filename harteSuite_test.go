package iz6502

/*
	Tests from https://github.com/TomHarte/ProcessorTests

	Know issues:
		- Test 6502/v1/20_55_13 (Note 1)
		- Not implemented undocumented opcodes for NMOS (Note 2)
		- Errors on flag N for ADC in BCD mode (Note 3)
		- Test 6502/v1/d3_f4_44 for undocumented opcode DCP (Note 4)

	The tests are disabled by default because they take long to run
	and require a huge download.
	To enable them, clone the repo https://github.com/SingleStepTests/65x02
	and change the variables ProcessorTestsEnable and ProcessorTestsPath.
*/

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var ProcessorTestsEnable = false
var ProcessorTestsPath = "../65x02-tests/"

type scenarioState struct {
	Pc  uint16
	S   uint8
	A   uint8
	X   uint8
	Y   uint8
	P   uint8
	Ram [][]uint16
}

type cycleEntry struct {
	Address   uint16
	Value     uint8
	Operation string
}

type scenario struct {
	Name    string
	Initial scenarioState
	Final   scenarioState
	Cycles  []cycleEntry
}

func TestHarteNMOS6502(t *testing.T) {
	if !ProcessorTestsEnable {
		t.Skip("TomHarte/ProcessorTests are not enabled")
	}

	s := NewNMOS6502(nil) // Use to get the opcodes names

	path := ProcessorTestsPath + "6502/v1/"
	for i := 0x00; i <= 0xff; i++ {
		mnemonic := s.opcodes[i].name
		if mnemonic != "" { // Note 2
			opcode := fmt.Sprintf("%02x", i)
			t.Run(opcode+mnemonic, func(t *testing.T) {
				t.Parallel()
				m := new(FlatMemory)
				s := NewNMOS6502(m)
				testOpcode(t, s, path, opcode, mnemonic)
			})
		}
	}
}

func TestHarteCMOS65c02(t *testing.T) {
	if !ProcessorTestsEnable {
		t.Skip("TomHarte/ProcessorTests are not enabled")
	}

	s := NewCMOS65c02(nil) // Use to get the opcodes names

	path := ProcessorTestsPath + "wdc65c02/v1/"
	for i := 0x00; i <= 0xff; i++ {
		mnemonic := s.opcodes[i].name
		opcode := fmt.Sprintf("%02x", i)
		t.Run(opcode+mnemonic, func(t *testing.T) {
			t.Parallel()
			m := new(FlatMemory)
			s := NewCMOS65c02(m)
			testOpcode(t, s, path, opcode, mnemonic)
		})
	}
}

func testOpcode(t *testing.T, s *State, path string, opcode string, mnemonic string) {
	data, err := os.ReadFile(path + opcode + ".json")
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		return
	}

	var scenarios []scenario
	err = json.Unmarshal(data, &scenarios)
	if err != nil {
		t.Fatal(err)
	}

	for _, scenario := range scenarios {
		if scenario.Name != "20 55 13" && // Note 1
			scenario.Name != "d3 f4 44" { // Note 4

			t.Run(scenario.Name, func(t *testing.T) {
				testScenario(t, s, &scenario, mnemonic)
			})
		}
	}
}

func testScenario(t *testing.T, s *State, sc *scenario, mnemonic string) {
	// Setup CPU
	start := s.GetCycles()
	s.reg.setPC(sc.Initial.Pc)
	s.reg.setSP(sc.Initial.S)
	s.reg.setA(sc.Initial.A)
	s.reg.setX(sc.Initial.X)
	s.reg.setY(sc.Initial.Y)
	s.reg.setP(sc.Initial.P)

	for _, e := range sc.Initial.Ram {
		s.mem.Poke(uint16(e[0]), uint8(e[1]))
	}

	// Execute instruction
	s.ExecuteInstruction()

	// Check result
	assertReg8(t, sc, "A", s.reg.getA(), sc.Final.A)
	assertReg8(t, sc, "X", s.reg.getX(), sc.Final.X)
	assertReg8(t, sc, "Y", s.reg.getY(), sc.Final.Y)
	if s.reg.getFlag(flagD) && (mnemonic == "ADC") {
		// Note 3
		assertFlags(t, sc, sc.Initial.P, s.reg.getP()&0x7f, sc.Final.P&0x7f)
	} else {
		assertFlags(t, sc, sc.Initial.P, s.reg.getP(), sc.Final.P)
	}
	assertReg8(t, sc, "SP", s.reg.getSP(), sc.Final.S)
	assertReg16(t, sc, "PC", s.reg.getPC(), sc.Final.Pc)

	cycles := s.GetCycles() - start
	if cycles != uint64(len(sc.Cycles)) {
		t.Errorf("Took %v cycles, it should be %v for %+v", cycles, len(sc.Cycles), sc)
	}
}

func assertReg8(t *testing.T, sc *scenario, name string, actual uint8, wanted uint8) {
	if actual != wanted {
		t.Errorf("Register %s is $%02x and should be $%02x for %+v", name, actual, wanted, sc)
	}
}

func assertReg16(t *testing.T, sc *scenario, name string, actual uint16, wanted uint16) {
	if actual != wanted {
		t.Errorf("Register %s is $%04x and should be $%04x for %+v", name, actual, wanted, sc)
	}
}

func assertFlags(t *testing.T, sc *scenario, initial uint8, actual uint8, wanted uint8) {
	if actual != wanted {
		t.Errorf("%08b flag diffs, they are %08b and should be %08b, initial %08b for %+v", actual^wanted, actual, wanted, initial, sc)
	}
}

func (c cycleEntry) String() string {
	if c.Operation == "read" {
		return fmt.Sprintf("[$%04X]->$%02X", c.Address, c.Value)
	} else if c.Operation == "write" {
		return fmt.Sprintf("[$%04X]<-$%02X", c.Address, c.Value)
	} else {
		return fmt.Sprintf("[$%04X](%s)$%02X ", c.Address, c.Operation, c.Value)
	}
}

func (s scenario) String() string {
	result := fmt.Sprintf("Name: %s\nInitial: %s\nFinal: %s\nCycles: ", s.Name, s.Initial.String(), s.Final.String())
	for _, c := range s.Cycles {
		result += fmt.Sprintf("%s ", c.String())
	}
	return result
}

func (s scenarioState) String() string {
	result := fmt.Sprintf("PC: $%04X, S: $%02X, A: $%02X, X: $%02X, Y: $%02X, P: $%02X, ", s.Pc, s.S, s.A, s.X, s.Y, s.P)
	result += "RAM:"
	for _, e := range s.Ram {
		if len(e) == 2 {
			result += fmt.Sprintf("  [$%04X]=$%02X", e[0], e[1])
		}
	}
	return result
}

func (s *scenario) UnmarshalJSON(data []byte) error {
	type Alias scenario
	aux := &struct {
		Cycles [][]interface{} `json:"cycles"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.Cycles = make([]cycleEntry, len(aux.Cycles))
	for i, c := range aux.Cycles {
		if len(c) != 3 {
			return fmt.Errorf("cycle entry does not have 3 elements: %v", c)
		}
		addr, _ := c[0].(float64)
		val, _ := c[1].(float64)
		op, _ := c[2].(string)
		s.Cycles[i] = cycleEntry{Address: uint16(addr), Value: uint8(val), Operation: op}
	}
	return nil
}

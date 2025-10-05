package vm

type Register interface {
	Operand
	isRegister()
}

type SpecialRegister int

const (
	_ SpecialRegister = iota
	PC
	BP
	SP
	HP
)

func (s SpecialRegister) isCode() {}
func (s SpecialRegister) String() string {
	return []string{
		PC: "pc",
		BP: "bp",
		SP: "sp",
		HP: "hp",
	}[s]
}
func (s SpecialRegister) isOperand()  {}
func (s SpecialRegister) isRegister() {}

type GeneralPurposeRegister int

const (
	_ GeneralPurposeRegister = iota
	R0
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
	R9
	R10
)

func (g GeneralPurposeRegister) isCode() {}
func (g GeneralPurposeRegister) String() string {
	return []string{
		R0:  "r0",
		R1:  "r1",
		R2:  "r2",
		R3:  "r3",
		R4:  "r4",
		R5:  "r5",
		R6:  "r6",
		R7:  "r7",
		R8:  "r8",
		R9:  "r9",
		R10: "r10",
	}[g]
}
func (g GeneralPurposeRegister) isOperand()  {}
func (g GeneralPurposeRegister) isRegister() {}

type FlagRegister int

const (
	_ FlagRegister = iota
	ZF
)

func (f FlagRegister) isCode() {}
func (f FlagRegister) String() string {
	return []string{
		ZF: "zf",
	}[f]
}
func (f FlagRegister) isOperand()  {}
func (f FlagRegister) isRegister() {}

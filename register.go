package minivm

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
	RA
	RB
	RC
	RD
	RE
)

func (g GeneralPurposeRegister) isCode() {}
func (g GeneralPurposeRegister) String() string {
	return []string{
		R0: "r0",
		R1: "r1",
		R2: "r2",
		R3: "r3",
		RA: "ra",
		RB: "rb",
		RC: "rc",
		RD: "rd",
		RE: "re",
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

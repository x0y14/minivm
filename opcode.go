package minivm

type Opcode int

const (
	NOP Opcode = iota

	MOV

	PUSH
	POP

	ALLOC
	STORE
	LOAD

	CALL
	RET

	JMP
	JZ
	JNZ

	ADD
	SUB

	EQ
	NE
	LT
	LE

	SYSCALL
)

func (o Opcode) isCode() {}

func (o Opcode) String() string {
	return []string{
		NOP:     "nop",
		MOV:     "mov",
		PUSH:    "push",
		POP:     "pop",
		ALLOC:   "alloc",
		STORE:   "store",
		LOAD:    "load",
		CALL:    "call",
		RET:     "ret",
		JMP:     "jmp",
		JZ:      "jz",
		JNZ:     "jnz",
		ADD:     "add",
		SUB:     "sub",
		EQ:      "eq",
		NE:      "ne",
		LT:      "lt",
		LE:      "le",
		SYSCALL: "syscall",
	}[o]
}

func (o Opcode) NumOperands() int {
	return []int{
		NOP:     0,
		MOV:     2,
		PUSH:    1,
		POP:     1,
		ALLOC:   1,
		STORE:   2,
		LOAD:    2,
		CALL:    1,
		RET:     0,
		JMP:     1,
		JZ:      1,
		JNZ:     1,
		ADD:     2,
		SUB:     2,
		EQ:      2,
		NE:      2,
		LT:      2,
		LE:      2,
		SYSCALL: 0,
	}[o]
}

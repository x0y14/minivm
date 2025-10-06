package asm

func isOperation(ident string) (Operation, bool) {
	switch ident {
	case "nop":
		return NOP, true
	case "mov":
		return MOV, true
	case "push":
		return PUSH, true
	case "pop":
		return POP, true
	case "alloc":
		return ALLOC, true
	case "store":
		return STORE, true
	case "load":
		return LOAD, true
	case "call":
		return CALL, true
	case "ret":
		return RET, true
	case "jmp":
		return JMP, true
	case "jz":
		return JZ, true
	case "jnz":
		return JNZ, true
	case "add":
		return ADD, true
	case "sub":
		return SUB, true
	case "eq":
		return EQ, true
	case "ne":
		return NE, true
	case "lt":
		return LT, true
	case "le":
		return LE, true
	case "syscall":
		return SYSCALL, true
	default:
		return 0, false
	}
}

func isRegister(ident string) (Register, bool) {
	switch ident {
	case "pc":
		return PC, true
	case "sp":
		return SP, true
	case "bp":
		return BP, true
	case "hp":
		return HP, true
	case "r0":
		return R0, true
	case "r1":
		return R1, true
	case "r2":
		return R2, true
	case "r3":
		return R3, true
	case "r4":
		return R4, true
	case "r5":
		return R5, true
	case "r6":
		return R6, true
	case "r7":
		return R7, true
	case "r8":
		return R8, true
	case "r9":
		return R9, true
	case "r10":
		return R10, true
	case "zf":
		return ZF, true
	default:
		return 0, false
	}
}

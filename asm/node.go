package asm

import (
	"fmt"
	"strconv"
	"strings"
)

type Node interface {
	isNode()
	String() string
}

type Operation int

const (
	_ Operation = iota
	NOP
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

func (o Operation) isNode() {}
func (o Operation) String() string {
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

// Instruction `mov dst src`のような命令
type Instruction struct {
	Op   Operation
	Args []Node
}

func (i Instruction) isNode() {}
func (i Instruction) String() string {
	var elms []string
	for _, elm := range i.Args {
		elms = append(elms, elm.String())
	}
	return fmt.Sprintf("Instruction{ Op: %s, Args: [ %s ] }", i.Op.String(), strings.Join(elms, ", "))
}

// Register `[sp+1]`の `sp`部分
type Register int

const (
	_ Register = iota
	PC
	SP
	BP
	HP

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

	ZF
)

func (r Register) isNode() {}
func (r Register) String() string {
	return []string{
		PC:  "pc",
		SP:  "sp",
		BP:  "bp",
		HP:  "hp",
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
		ZF:  "zf",
	}[r]
}

// Offset `[sp+1]`のような相対位置
type Offset struct {
	Target Register
	Diff   int
}

func (o Offset) isNode() {}
func (o Offset) String() string {
	return fmt.Sprintf("[%s%+d]", o.Target.String(), o.Diff)
}

type Number int

func (n Number) isNode() {}
func (n Number) String() string {
	return strconv.Itoa(int(n))
}

type Character rune

func (c Character) isNode() {}
func (c Character) String() string {
	return strconv.QuoteRune(rune(c))
}

func Parse(token *Token) ([]Instruction, error) {
	return nil, nil
}

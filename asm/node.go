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

var curt *Token

func expect(kind TokenKind) (*Token, error) {
	if curt.Kind != kind {
		return nil, fmt.Errorf("want=%s, got=%s", kind.String(), curt.Kind.String())
	}
	v := *curt
	curt = curt.Next
	return &v, nil
}
func consume(kind TokenKind) *Token {
	if curt.Kind != kind {
		return nil
	}
	v := *curt
	curt = curt.Next
	return &v
}

func parsePcOffset() ([]Node, error) {
	// (
	if _, err := expect(Lrb); err != nil {
		return nil, err
	}

	// +
	plus := consume(Add)
	// -
	minus := consume(Sub)
	if plus != nil && minus != nil {
		return nil, fmt.Errorf("syntax err")
	}

	// diff
	diff, err := expect(Integer)
	if err != nil {
		return nil, err
	}

	// )
	if _, err = expect(Rrb); err != nil {
		return nil, err
	}

	v, err := diff.GetValueAsInteger()
	if err != nil {
		return nil, err
	}
	if minus != nil {
		return []Node{Offset{PC, -v}}, nil
	}
	return []Node{Offset{PC, v}}, nil
}

func parseStackOffset() ([]Node, error) {
	// [
	if _, err := expect(Lcb); err != nil {
		return nil, err
	}

	// sp / bp
	id, err := expect(Identifier)
	if err != nil {
		return nil, err
	}
	var reg Register
	switch string(id.Raw) {
	case "sp":
		reg = SP
	case "bp":
		reg = BP
	default:
		return nil, fmt.Errorf("unsupported register: %s", string(id.Raw))
	}

	// +
	plus := consume(Add)
	// -
	minus := consume(Sub)
	if plus != nil && minus != nil {
		return nil, fmt.Errorf("syntax err")
	}

	// diff
	diff, err := expect(Integer)
	if err != nil {
		return nil, err
	}

	// ]
	if _, err := expect(Rcb); err != nil {
		return nil, err
	}

	v, err := diff.GetValueAsInteger()
	if err != nil {
		return nil, err
	}
	if minus != nil {
		return []Node{Offset{reg, -v}}, nil
	}
	return []Node{Offset{reg, v}}, nil
}

func Parse(token *Token) ([]Node, error) {
	var nodes []Node
	curt = token
loop:
	for {
		switch curt.Kind {
		case Eof:
			break loop
		case Comment:
			curt = curt.Next
		case Identifier:
			if op, yes := isOperation(string(curt.Raw)); yes {
				nodes = append(nodes, op)
				curt = curt.Next
				continue
			}
			if reg, yes := isRegister(string(curt.Raw)); yes {
				nodes = append(nodes, reg)
				curt = curt.Next
				continue
			}
			return nil, fmt.Errorf("parse: unsupported ident: %s", string(curt.Raw))
		case Integer:
			v, err := curt.GetValueAsInteger()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, Number(v))
			curt = curt.Next
		case Char:
			v, err := curt.GetValueAsRune()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, Character(v))
			curt = curt.Next
		case Lrb:
			nds, err := parsePcOffset()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, nds...)
		case Lcb:
			nds, err := parseStackOffset()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, nds...)
		default:
			return nil, fmt.Errorf("parse: unsupported token: %s", curt.Kind.String())
		}
	}
	return nodes, nil
}

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
)

func (o Operation) isNode() {}
func (o Operation) String() string {
	return []string{
		NOP: "nop",
		MOV: "mov",
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
)

func (r Register) isNode() {}
func (r Register) String() string {
	return []string{
		PC: "pc",
		SP: "sp",
		BP: "bp",
		HP: "hp",
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

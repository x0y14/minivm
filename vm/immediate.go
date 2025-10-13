package vm

import "strconv"

type Immediate interface {
	Operand
	isImmediate()
	Value() int
}

type Integer int

func (i Integer) isCode() {}
func (i Integer) String() string {
	return strconv.Itoa(int(i))
}
func (i Integer) isOperand()   {}
func (i Integer) isImmediate() {}
func (i Integer) Value() int {
	return int(i)
}

type Boolean bool

func (b Boolean) isCode() {}
func (b Boolean) String() string {
	if b {
		return "true"
	}
	return "false"
}
func (b Boolean) isOperand()   {}
func (b Boolean) isImmediate() {}
func (b Boolean) Value() int {
	if b {
		return 1
	}
	return 0
}

type Character rune

func (c Character) isCode() {}
func (c Character) String() string {
	if c == 0 {
		return "'\\0'"
	}
	return strconv.QuoteRune(rune(c))
}
func (c Character) isOperand()   {}
func (c Character) isImmediate() {}
func (c Character) Value() int {
	return int(c)
}

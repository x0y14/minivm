package vm

import (
	"fmt"
)

type Offset interface {
	Operand
	isOffset()
}

type BpOffset int

func (b BpOffset) isCode() {}
func (b BpOffset) String() string {
	return fmt.Sprintf("[%s%+d]", BP.String(), b)
}
func (b BpOffset) isOperand() {}
func (b BpOffset) isOffset()  {}

type SpOffset int

func (s SpOffset) isCode() {}
func (s SpOffset) String() string {
	return fmt.Sprintf("[%s%+d]", SP.String(), s)
}
func (s SpOffset) isOperand() {}
func (s SpOffset) isOffset()  {}

type PcOffset int

func (p PcOffset) isCode() {}
func (p PcOffset) String() string {
	return fmt.Sprintf("(%+d)", p)
}
func (p PcOffset) isOperand() {}
func (p PcOffset) isOffset()  {}

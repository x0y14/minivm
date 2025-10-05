package vm

type Operand interface {
	Code
	isOperand()
}

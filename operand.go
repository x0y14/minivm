package minivm

type Operand interface {
	Code
	isOperand()
}

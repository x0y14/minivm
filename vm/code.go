package vm

type Code interface {
	isCode()
	String() string
}

package minivm

type Code interface {
	isCode()
	String() string
}

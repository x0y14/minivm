package vm

type Type int

const (
	TUndefined Type = iota
	TInt
	TBool
)

func typeof(imm Immediate) Type {
	switch imm.(type) {
	case Integer:
		return TInt
	case Boolean:
		return TBool
	default:
		return TUndefined
	}
}

func match(imm1, imm2 Immediate) bool {
	return typeof(imm1) == typeof(imm2)
}

func calculable(imm1, imm2 Immediate) bool {
	if !match(imm1, imm2) {
		return false
	}
	if typeof(imm1) != TInt {
		return false
	}
	return true
}

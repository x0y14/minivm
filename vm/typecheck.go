package vm

type Type int

const (
	TUndefined Type = iota
	TInt
	TBool
	TChar
)

func typeof(imm Immediate) Type {
	switch imm.(type) {
	case Integer:
		return TInt
	case Boolean:
		return TBool
	case Character:
		return TChar
	default:
		return TUndefined
	}
}

func match(imm1, imm2 Immediate) bool {
	return typeof(imm1) == typeof(imm2)
}

func calculable(imm1, imm2 Immediate) bool {
	switch {
	case match(imm1, imm2):
		return true
	case typeof(imm1) == TInt && typeof(imm2) == TChar:
		return true
	case typeof(imm1) == TChar && typeof(imm2) == TInt:
		return true
	default:
		return false
	}
}

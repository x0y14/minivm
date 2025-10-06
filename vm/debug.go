package vm

import (
	"fmt"
	"io"
)

// DumpText writes a human-readable listing of program using Code.String().
// Each instruction is emitted on one line based on the opcode's operand count.
func DumpText(w io.Writer, program []Code) error {
	for i := 0; i < len(program); {
		c := program[i]
		op, ok := c.(Opcode)
		if !ok {
			return fmt.Errorf("DumpText: program[%d] is not an opcode: %T", i, c)
		}
		n := op.NumOperands()
		if i+n >= len(program) {
			return fmt.Errorf("DumpText: opcode at %d expects %d operands, but program ends early", i, n)
		}

		// opcode
		if _, err := io.WriteString(w, op.String()); err != nil {
			return err
		}
		// operands
		for j := 1; j <= n; j++ {
			if _, err := io.WriteString(w, " "); err != nil {
				return err
			}
			if _, err := io.WriteString(w, program[i+j].String()); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}

		i += n + 1
	}
	return nil
}

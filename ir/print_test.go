package ir

import "testing"

func TestPrint_AlignZeroOperand(t *testing.T) {
	nodes := []Node{
		Instruction{Op: NOP, Args: []Node{}},
		Instruction{Op: SYSCALL, Args: []Node{}},
	}
	got := Print(expand(nodes))
	want := "nop\nsyscall\n"
	if got != want {
		t.Fatalf("unexpected output:\n got: %q\nwant: %q", got, want)
	}
}

func TestPrint_PadsWithinOperandGroup(t *testing.T) {
	nodes := []Node{
		Instruction{Op: PUSH, Args: []Node{R0}},
		Instruction{Op: ALLOC, Args: []Node{Number(16)}},
	}
	got := Print(expand(nodes))
	// PUSH (len=4) と ALLOC (len=5) はオペランド数=1 のグループなので
	// PUSH は ALLOC に合わせて 1 文字分パディングされる想定
	want := "push r0\nalloc 16\n"
	if got != want {
		t.Fatalf("unexpected output:\n got: %q\nwant: %q", got, want)
	}
}

func TestPrint_MixedInstructionsAndNumber(t *testing.T) {
	nodes := []Node{
		Instruction{Op: MOV, Args: []Node{R0, R1}},
		Instruction{Op: ADD, Args: []Node{R0, Number(2)}},
		Number(5),
	}
	got := Print(expand(nodes))
	want := "mov r0 r1\nadd r0 2\n5\n"
	if got != want {
		t.Fatalf("unexpected output:\n got: %q\nwant: %q", got, want)
	}
}

package asm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse_Basic(t *testing.T) {
	input := `
mov r0 1
add r1 r0
push 'a'
`
	toks, err := Tokenize([]rune(input))
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	nodes, err := Parse(toks)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	expect := []Node{
		MOV, R0, Number(1),
		ADD, R1, R0,
		PUSH, Character('a'),
	}
	if diff := cmp.Diff(expect, nodes); diff != "" {
		t.Errorf("diff:\n%s", diff)
	}
}

func TestParse_Offsets(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []Node
	}{
		{
			name:  "pc offset +",
			input: "jmp (+3)",
			expect: []Node{
				JMP, Offset{Target: PC, Diff: 3},
			},
		},
		{
			name:  "pc offset -",
			input: "jmp (-15)",
			expect: []Node{
				JMP, Offset{Target: PC, Diff: -15},
			},
		},
		{
			name:  "pc offset no sign",
			input: "jmp (7)",
			expect: []Node{
				JMP, Offset{Target: PC, Diff: 7},
			},
		},
		{
			name:  "stack offset sp +",
			input: "load r2 [sp+4]",
			expect: []Node{
				LOAD, R2, Offset{Target: SP, Diff: 4},
			},
		},
		{
			name:  "stack offset bp -",
			input: "store [bp-2] r3",
			expect: []Node{
				STORE, Offset{Target: BP, Diff: -2}, R3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks, err := Tokenize([]rune(tt.input))
			if err != nil {
				t.Fatalf("tokenize error: %v", err)
			}
			nodes, err := Parse(toks)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			if diff := cmp.Diff(tt.expect, nodes); diff != "" {
				t.Errorf("diff:\n%s", diff)
			}
		})
	}
}

func TestParse_SkipComment(t *testing.T) {
	input := "mov r0 1 ; comment\nadd r0 2"
	toks, err := Tokenize([]rune(input))
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	nodes, err := Parse(toks)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	expect := []Node{
		MOV, R0, Number(1),
		ADD, R0, Number(2),
	}
	if diff := cmp.Diff(expect, nodes); diff != "" {
		t.Errorf("diff:\n%s", diff)
	}
}

func TestParse_Error_UnsupportedIdent(t *testing.T) {
	input := "unknown"
	toks, err := Tokenize([]rune(input))
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	if _, err := Parse(toks); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestParse_Error_StackOffset_UnsupportedRegister(t *testing.T) {
	input := "[hp+1]"
	toks, err := Tokenize([]rune(input))
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	if _, err := Parse(toks); err == nil {
		t.Fatalf("expected error, got nil")
	}
}

package ir

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// AUTO 定数ラベルが load/store のオペランドから除去され、数値になること
func TestLink_ResolvesAutoDataLabelsInOperands(t *testing.T) {
	code := `
.section .data:
    num auto "A"

.section .text:
    global _start
_start:
    load num r1
    store num r1
`
	tokens, err := Tokenize([]rune(code), true)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := Link([]*IR{ir})
	if err != nil {
		t.Fatal(err)
	}

	// ラベル num が最終ノード列に存在しないこと
	for _, n := range nodes {
		if lb, ok := n.(Label); ok && lb.Name == "num" {
			t.Fatalf("label %q still remains in linked output", lb.Name)
		}
	}

	// load num r1 -> load <number> r1 に解決されていること
	foundLoad := false
	foundStore := false
	for i := 0; i+2 < len(nodes); i++ {
		if op, ok := nodes[i].(Operation); ok && op == LOAD {
			if _, ok := nodes[i+1].(Number); ok {
				if r, ok := nodes[i+2].(Register); ok && r == R1 {
					foundLoad = true
				}
			}
		}
		if op, ok := nodes[i].(Operation); ok && op == STORE {
			if _, ok := nodes[i+1].(Number); ok {
				if r, ok := nodes[i+2].(Register); ok && r == R1 {
					foundStore = true
				}
			}
		}
	}
	if !foundLoad {
		t.Fatalf("resolved LOAD <number> r1 not found")
	}
	if !foundStore {
		t.Fatalf("resolved STORE <number> r1 not found")
	}
}

// sizeof 定数参照がテキストで数値に解決されること
func TestLink_ReplacesSizeofInText(t *testing.T) {
	code := `
.section .data:
    arr auto "hi"
    sz sizeof arr

.section .text:
    global _start
_start:
    mov r1 sz
`
	tokens, err := Tokenize([]rune(code), true)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := Link([]*IR{ir})
	if err != nil {
		t.Fatal(err)
	}

	// mov r1 2 が存在すること（"hi" の長さ）
	found := false
	for i := 0; i+2 < len(nodes); i++ {
		if op, ok := nodes[i].(Operation); ok && op == MOV {
			if r, ok := nodes[i+1].(Register); ok && r == R1 {
				if n, ok := nodes[i+2].(Number); ok && int(n) == 2 {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Fatalf("MOV r1 2 (sizeof resolution) not found in linked output")
	}

	// ラベル sz が残っていないこと
	for _, n := range nodes {
		if lb, ok := n.(Label); ok && lb.Name == "sz" {
			t.Fatalf("label %q still remains in linked output", lb.Name)
		}
	}
}

// プリスクリプトに ALLOC と STORE が生成されること（AUTO データの初期化）
func TestLink_GeneratesPreScriptForAutoData(t *testing.T) {
	code := `
.section .data:
    data auto "AB"

.section .text:
    global _start
_start:
    nop
`
	tokens, err := Tokenize([]rune(code), true)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	nodes, err := Link([]*IR{ir})
	if err != nil {
		t.Fatal(err)
	}

	// プリスクリプト中に ALLOC 2, POP r10 があること
	foundAlloc := false
	for i := 0; i+3 < len(nodes); i++ {
		if op, ok := nodes[i].(Operation); ok && op == ALLOC {
			if n, ok := nodes[i+1].(Number); ok && int(n) == 2 {
				if op2, ok := nodes[i+2].(Operation); ok && op2 == POP {
					if r, ok := nodes[i+3].(Register); ok && r == R10 {
						foundAlloc = true
						break
					}
				}
			}
		}
	}
	if !foundAlloc {
		t.Fatalf("pre-script ALLOC 2; POP r10 not found")
	}

	// data[0]='A', data[1]='B' の STORE があること
	foundStoreA := false
	foundStoreB := false
	for i := 0; i+2 < len(nodes); i++ {
		if op, ok := nodes[i].(Operation); ok && op == STORE {
			if base, ok := nodes[i+1].(Number); ok {
				switch int(base) {
				case 0:
					if ch, ok := nodes[i+2].(Character); ok && rune(ch) == 'A' {
						foundStoreA = true
					}
				case 1:
					if ch, ok := nodes[i+2].(Character); ok && rune(ch) == 'B' {
						foundStoreB = true
					}
				}
			}
		}
	}
	if !foundStoreA || !foundStoreB {
		t.Fatalf("pre-script STORE for data bytes not found: A=%v B=%v", foundStoreA, foundStoreB)
	}
}

func TestLink_FizzBuzz(t *testing.T) {
	code := `
.section .data:
    ; number buffer as separate one-byte constants so テキスト中から個別に更新できる
    num0 auto "0"
    num1 auto "0"
    num2 auto "1"
    num3 auto "\n"

    fizz auto "Fizz\n"
    buzz auto "Buzz\n"
    fizzbuzz auto "FizzBuzz\n"

.section .text:
    global _start

_start:
    ; r5 = current number (1..100)
    mov r5 1
    ; r7 = counter mod3, r8 = counter mod5
    mov r7 0
    mov r8 0

loop_start:
    ; --- update mod3 ---
    add r7 1
    eq r7 3
    jz reset_mod3
    mov r9 0        ; fizzFlag = 0
    jmp after_mod3
reset_mod3:
    mov r7 0
    mov r9 1        ; fizzFlag = 1
after_mod3:

    ; --- update mod5 ---
    add r8 1
    eq r8 5
    jz reset_mod5
    mov r10 0       ; buzzFlag = 0
    jmp after_mod5
reset_mod5:
    mov r8 0
    mov r10 1       ; buzzFlag = 1
after_mod5:

    ; decide what to print
    mov r4 r9
    add r4 r10      ; r4 = fizzFlag + buzzFlag
    eq r4 2
    jz print_fizzbuzz
    eq r4 1
    jz print_fizz_or_buzz

print_number:
    ; print zero-padded 3-digit number + newline at num0 (num0,num1,num2,num3)
    mov r1 1
    mov r2 num0
    mov r3 4
    mov r0 1
    syscall
    jmp after_print

print_fizz_or_buzz:
    eq r9 1
    jz print_fizz
    ; else buzz
print_buzz:
    mov r1 1
    mov r2 buzz
    mov r3 5
    mov r0 1
    syscall
    jmp after_print
print_fizz:
    mov r1 1
    mov r2 fizz
    mov r3 5
    mov r0 1
    syscall
    jmp after_print

print_fizzbuzz:
    mov r1 1
    mov r2 fizzbuzz
    mov r3 9
    mov r0 1
    syscall

after_print:
    ; --- increment number buffer stored in heap as num2 (ones), num1 (tens), num0 (hundreds) ---
    ; load num2 -> r1, add 1, if overflow ('9'+1 => 58) carry else store back
    load num2 r1
    add r1 1
    eq r1 58
    jz carry_ones
    store num2 r1
    jmp cont_inc
carry_ones:
    mov r1 48        ; '0'
    store num2 r1
    ; carry to tens
    load num1 r1
    add r1 1
    eq r1 58
    jz carry_tens
    store num1 r1
    jmp cont_inc
carry_tens:
    mov r1 48
    store num1 r1
    ; carry to hundreds (no wrap for 100)
    load num0 r1
    add r1 1
    store num0 r1

cont_inc:
    ; increment loop counter
    add r5 1
    ; stop when r5 == 101
    eq r5 101
    jz program_done
    jmp loop_start

program_done:
    ; exit syscall (R0=0, R1=status)
    mov r1 0
    mov r0 0
    syscall
`
	tokens, err := Tokenize([]rune(code), true)
	if err != nil {
		t.Fatal(err)
	}
	ir, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Link([]*IR{ir})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLink_RelativeOffsetCalculation(t *testing.T) {
	tests := []struct {
		name   string
		irs    []*IR
		expect []Node
	}{
		{
			name: "forward jump",
			irs: []*IR{
				{
					EntryPoint: "_start",
					Text: expand([]Node{
						Label{Define: true, Name: "_start"},
						Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_target"}}},
						NOP,
						NOP,
						Label{Define: true, Name: "_target"},
						Instruction{Op: MOV, Args: []Node{R0, Number(0)}},
					}),
				},
			},
			expect: []Node{
				JMP, Offset{PC, 11}, // goto _pre
				NOP,                // _start
				JMP, Offset{PC, 4}, // forward +4
				NOP,
				NOP,
				NOP, // _target
				MOV, R0, Number(0),
				NOP,                  // _pre
				JMP, Offset{PC, -10}, // goto _start
			},
		},
		{
			name: "backward jump",
			irs: []*IR{
				{
					EntryPoint: "_start",
					Text: expand([]Node{
						Label{Define: true, Name: "_start"},
						Label{Define: true, Name: "_loop"},
						Instruction{Op: SUB, Args: []Node{R1, Number(1)}},
						Instruction{Op: JNZ, Args: []Node{Label{Define: false, Name: "_loop"}}},
					}),
				},
			},
			expect: []Node{
				JMP, Offset{PC, +9}, // goto _pre
				NOP, // _start
				NOP, // _loop
				SUB, R1, Number(1),
				JNZ, Offset{PC, -4}, // backward -4 (goto _loop)
				NOP,                 // _pre
				JMP, Offset{PC, -8}, // goto _start
			},
		},
		{
			name: "call with offset",
			irs: []*IR{
				{
					EntryPoint: "_start",
					Text: expand([]Node{
						Label{Define: true, Name: "_start"},
						Instruction{Op: CALL, Args: []Node{Label{Define: false, Name: "_func"}}},
						Instruction{Op: RET, Args: []Node{}},
						Label{Define: true, Name: "_func"},
						Instruction{Op: MOV, Args: []Node{R0, Number(42)}},
						Instruction{Op: RET, Args: []Node{}},
					}),
				},
			},
			expect: []Node{
				JMP, Offset{PC, 11}, // goto _pre
				NOP,                 // _start
				CALL, Offset{PC, 3}, // goto _func
				RET,
				NOP, // _func
				MOV, R0, Number(42),
				RET,
				NOP,                  // _pre
				JMP, Offset{PC, -10}, // goto _start
			},
		},
		{
			name: "multiple jumps same target",
			irs: []*IR{
				{
					EntryPoint: "_start",
					Text: expand([]Node{
						Label{Define: true, Name: "_start"},
						Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_end"}}},
						NOP,
						Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_end"}}},
						NOP,
						Label{Define: true, Name: "_end"},
						Instruction{Op: RET, Args: []Node{}},
					}),
				},
			},
			expect: []Node{
				JMP, Offset{PC, 11}, // goto _pre
				NOP,                // _start
				JMP, Offset{PC, 6}, // goto _end
				NOP,
				JMP, Offset{PC, 3}, // goto _end
				NOP,
				NOP, // _end
				RET,
				NOP,                  // _pre
				JMP, Offset{PC, -10}, // _start
			},
		},
		{
			name: "cross-module label reference",
			irs: []*IR{
				{
					Exports:    []string{"_func"},
					EntryPoint: "",
					Text: expand([]Node{
						Label{Define: true, Name: "_func"},
						Instruction{Op: MOV, Args: []Node{R0, Number(1)}},
						Instruction{Op: RET, Args: []Node{}},
					}),
				},
				{
					Imports:    []string{"_func"},
					EntryPoint: "_start",
					Text: expand([]Node{
						Label{Define: true, Name: "_start"},
						Instruction{Op: CALL, Args: []Node{Label{Define: false, Name: "_func"}}},
						Instruction{Op: RET, Args: []Node{}},
					}),
				},
			},
			expect: []Node{
				JMP, Offset{PC, 11}, // goto _pre
				NOP, // _func
				MOV, R0, Number(1),
				RET,
				NOP,                  // _start
				CALL, Offset{PC, -6}, // goto _func
				RET,
				NOP,                 // _pre
				JMP, Offset{PC, -5}, // goto _start
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Link(tt.irs)
			if err != nil {
				t.Fatalf("Link error: %v", err)
			}

			if diff := cmp.Diff(tt.expect, got); diff != "" {
				t.Errorf("diff (-want +got):\n%s", diff)
			}
		})
	}
}

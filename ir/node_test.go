package ir

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse_Sample1(t *testing.T) {
	input := strings.TrimSpace(`
.import printf
.export _print_fizz

.section .data:
    msg auto "hello"
	msgLen sizeof msg
    arr auto 10, 20, 30

.section .text:
    global _start

_start:
    alloc 16
    pop r10
    mov r6 1
`)

	toks, err := Tokenize([]rune(input), true)
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}
	irObj, err := Parse(toks)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// imports に printf があること
	found := false
	for _, im := range irObj.Imports {
		if im == "printf" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("imports does not contain printf: %v", irObj.Imports)
	}

	// exports に _print_fizz があること
	found = false
	for _, ex := range irObj.Exports {
		if ex == "_print_fizz" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("exports does not contain _print_fizz: %v", irObj.Exports)
	}

	// constants の msg が AUTO かつ 値が "hello" であること
	cst, ok := irObj.Constants["msg"]
	if !ok {
		t.Fatalf("constants does not contain msg")
	}
	if cst.Mode != AUTO {
		t.Fatalf("msg mode want=AUTO got=%v", cst.Mode)
	}
	exp := "hello"
	if len(cst.Values) != len(exp) {
		t.Fatalf("msg values length want=%d got=%d", len(exp), len(cst.Values))
	}
	for i, r := range exp {
		if cc, ok := cst.Values[i].(ConstChar); !ok || string(cc) != string(r) {
			t.Fatalf("msg value[%d] want=%q got=%#v", i, string(r), cst.Values[i])
		}
	}

	// entrypoint が _start であること
	if irObj.EntryPoint != "_start" {
		t.Fatalf("entrypoint want=_start got=%q", irObj.EntryPoint)
	}

	// プログラム中に定義ラベル _start があること
	foundLabel := false
	for _, n := range irObj.Text {
		if lb, ok := n.(Label); ok {
			if lb.Name == "_start" && lb.Define {
				foundLabel = true
				break
			}
		}
	}
	if !foundLabel {
		t.Fatalf("text does not contain defined label _start")
	}
}

func expand(nodes []Node) []Node {
	var nds []Node
	for _, n := range nodes {
		switch n := n.(type) {
		case Instruction:
			nds = append(nds, n.Nodes()...)
		default:
			nds = append(nds, n)
		}
	}
	return nds
}

func TestParse_Calc(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		expect *IR
	}{
		{
			"lib",
			`
.export _add
.export _sub
.export _mul

.section .text:
; add: r0 = r1 + r2
_add:
    mov r0 r1
    add r0 r2
    jmp _lib_ret

; sub: r0 = r1 - r2
_sub:
    mov r0 r1
    sub r0 r2
    jmp _lib_ret

; mul: r0 = r1 * r2  (繰り返し加算の簡易実装)
_mul:
    mov r0 0        ; result = 0
    mov r3 r2       ; counter = r2
_mul_loop:
    eq r3 0
    jz _mul_done
    add r0 r1
    sub r3 1
    jmp _mul_loop
_mul_done:
    jmp _lib_ret

; ライブラリ共通の戻りプレースホルダ
_lib_ret:
    ; return placeholder`,
			&IR{
				Imports:    []string{},
				Exports:    []string{"_add", "_sub", "_mul"},
				Constants:  map[string]Constant{},
				EntryPoint: "",
				Text: expand([]Node{
					Label{Define: true, Name: "_add"},
					Instruction{Op: MOV, Args: []Node{R0, R1}},
					Instruction{Op: ADD, Args: []Node{R0, R2}},
					Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_lib_ret"}}},

					Label{Define: true, Name: "_sub"},
					Instruction{Op: MOV, Args: []Node{R0, R1}},
					Instruction{Op: SUB, Args: []Node{R0, R2}},
					Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_lib_ret"}}},

					Label{Define: true, Name: "_mul"},
					Instruction{Op: MOV, Args: []Node{R0, Number(0)}},
					Instruction{Op: MOV, Args: []Node{R3, R2}},
					Label{Define: true, Name: "_mul_loop"},
					Instruction{Op: EQ, Args: []Node{R3, Number(0)}},
					Instruction{Op: JZ, Args: []Node{Label{Define: false, Name: "_mul_done"}}},
					Instruction{Op: ADD, Args: []Node{R0, R1}},
					Instruction{Op: SUB, Args: []Node{R3, Number(1)}},
					Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_mul_loop"}}},
					Label{Define: true, Name: "_mul_done"},
					Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_lib_ret"}}},

					Label{Define: true, Name: "_lib_ret"},
				}),
			},
		},
		{
			"main",
			`
.import add
.import sub
.import mul

.section .text:
    global _start

; main(): return mul( add(1,2), sub(4,3) )
_start:
    ; add(1,2)
    mov r1 1
    mov r2 2
    jmp _add
_after_add:
    mov r3 r0       ; add の結果を一時保存 (r3)

    ; sub(4,3)
    mov r1 4
    mov r2 3
    jmp _sub
_after_sub:
    mov r2 r0       ; sub の結果を r2 (mul の第2引数)
    mov r1 r3       ; add の結果を r1 (mul の第1引数)
    jmp _mul
_after_mul:
    mov r1 r0       ; 計算結果をr1へ
    mov r0 0        ; sys_exit
    syscall`,
			&IR{
				Imports:    []string{"add", "sub", "mul"},
				Exports:    []string{},
				Constants:  map[string]Constant{},
				EntryPoint: "_start",
				Text: expand([]Node{
					Label{Define: true, Name: "_start"},
					Instruction{Op: MOV, Args: []Node{R1, Number(1)}},
					Instruction{Op: MOV, Args: []Node{R2, Number(2)}},
					Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_add"}}},

					Label{Define: true, Name: "_after_add"},
					Instruction{Op: MOV, Args: []Node{R3, R0}},

					Instruction{Op: MOV, Args: []Node{R1, Number(4)}},
					Instruction{Op: MOV, Args: []Node{R2, Number(3)}},
					Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_sub"}}},

					Label{Define: true, Name: "_after_sub"},
					Instruction{Op: MOV, Args: []Node{R2, R0}},
					Instruction{Op: MOV, Args: []Node{R1, R3}},
					Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "_mul"}}},

					Label{Define: true, Name: "_after_mul"},
					Instruction{Op: MOV, Args: []Node{R1, R0}},
					Instruction{Op: MOV, Args: []Node{R0, Number(0)}},
					Instruction{Op: SYSCALL, Args: []Node{}},
				}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := Tokenize([]rune(tt.in), true)
			if err != nil {
				t.Fatal(err)
			}
			ir, err := Parse(tok)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expect, ir); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

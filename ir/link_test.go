package ir

import (
	"testing"
)

//func containsString(slice []string, s string) bool {
//	for _, v := range slice {
//		if v == s {
//			return true
//		}
//	}
//	return false
//}
//
//func findOffsets(ir *IR) []Offset {
//	var res []Offset
//	for _, n := range ir.Text {
//		if o, ok := n.(Offset); ok {
//			res = append(res, o)
//		}
//		// also check inside Instruction args
//		if ins, ok := n.(Instruction); ok {
//			for _, a := range ins.Args {
//				if o, ok := a.(Offset); ok {
//					res = append(res, o)
//				}
//			}
//		}
//	}
//	return res
//}
//
//func TestLink_MergeRemovesImportAndResolves(t *testing.T) {
//	// main imports "foo" and jumps to it; lib defines "foo"
//	main := &IR{
//		Imports:    []string{"foo"},
//		Exports:    []string{},
//		Constants:  []Constant{},
//		EntryPoint: "_start",
//		Text: []Node{
//			Label{Define: true, Name: "_start"},
//			Instruction{Op: JMP, Args: []Node{Label{Define: false, Name: "foo"}}},
//		},
//	}
//	lib := &IR{
//		Imports:    []string{},
//		Exports:    []string{"foo"},
//		Constants:  []Constant{},
//		EntryPoint: "",
//		Text: []Node{
//			Label{Define: true, Name: "foo"},
//			NOP,
//		},
//	}
//
//	res, err := Link([]*IR{main, lib})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//	// import "foo" は解決されていること
//	if containsString(res.Imports, "foo") {
//		t.Fatalf("import foo should be removed after linking, got imports=%v", res.Imports)
//	}
//	// entrypoint が引き継がれていること
//	if res.EntryPoint != "_start" {
//		t.Fatalf("EntryPoint expected _start, got=%q", res.EntryPoint)
//	}
//}
//
//func TestLink_TooManyEntryPoints(t *testing.T) {
//	a := &IR{EntryPoint: "_one", Imports: []string{}, Text: []Node{}}
//	b := &IR{EntryPoint: "_two", Imports: []string{}, Text: []Node{}}
//	_, err := Link([]*IR{a, b})
//	if err == nil {
//		t.Fatalf("expected error when multiple entry points exist")
//	}
//	if !strings.Contains(err.Error(), "too many entryPoint") {
//		t.Fatalf("unexpected error: %v", err)
//	}
//}
//
//func TestLink_UnsolvedLabel_Error(t *testing.T) {
//	// 未定義ラベルを参照しているが import でも定義でもない -> エラー
//	ir := &IR{
//		Imports:    []string{},
//		Exports:    []string{},
//		Constants:  []Constant{},
//		EntryPoint: "",
//		Text: []Node{
//			Label{Define: false, Name: "missing"},
//		},
//	}
//	_, err := Link([]*IR{ir})
//	if err == nil {
//		t.Fatalf("expected unsolved label error")
//	}
//	if !strings.Contains(err.Error(), "undefined") {
//		t.Fatalf("unexpected error: %v", err)
//	}
//}
//
//func TestLink_OffsetAdjustmentOnMerge(t *testing.T) {
//	// dst has 3 nodes, src has an Offset{PC,5} which should become 5+3=8 after merge
//	dst := &IR{
//		Imports:    []string{},
//		Exports:    []string{},
//		Constants:  []Constant{},
//		EntryPoint: "",
//		Text: []Node{
//			NOP, NOP, NOP, // length 3
//		},
//	}
//	src := &IR{
//		Imports:    []string{},
//		Exports:    []string{},
//		Constants:  []Constant{},
//		EntryPoint: "",
//		Text: []Node{
//			Offset{Target: PC, Diff: 5},
//		},
//	}
//
//	res, err := Link([]*IR{dst, src})
//	if err != nil {
//		t.Fatalf("unexpected error: %v", err)
//	}
//	offs := findOffsets(res)
//	if len(offs) == 0 {
//		t.Fatalf("expected at least one Offset in linked IR")
//	}
//	found := false
//	for _, o := range offs {
//		if o.Target == PC && o.Diff == 8 {
//			found = true
//			break
//		}
//	}
//	if !found {
//		t.Fatalf("expected adjusted Offset with Diff=8, got offsets=%v", offs)
//	}
//}

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

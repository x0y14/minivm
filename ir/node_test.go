package ir

import (
	"strings"
	"testing"
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

	toks, err := Tokenize([]rune(input))
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

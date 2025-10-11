package ir

import (
	"fmt"
	"strconv"
	"strings"
)

type Node interface {
	isNode()
	String() string
}

type Operation int

const (
	_ Operation = iota
	NOP
	MOV
	PUSH
	POP
	ALLOC
	STORE
	LOAD
	CALL
	RET
	JMP
	JZ
	JNZ
	ADD
	SUB
	EQ
	NE
	LT
	LE
	SYSCALL
)

func (o Operation) isNode() {}
func (o Operation) String() string {
	return []string{
		NOP:     "nop",
		MOV:     "mov",
		PUSH:    "push",
		POP:     "pop",
		ALLOC:   "alloc",
		STORE:   "store",
		LOAD:    "load",
		CALL:    "call",
		RET:     "ret",
		JMP:     "jmp",
		JZ:      "jz",
		JNZ:     "jnz",
		ADD:     "add",
		SUB:     "sub",
		EQ:      "eq",
		NE:      "ne",
		LT:      "lt",
		LE:      "le",
		SYSCALL: "syscall",
	}[o]
}

// Instruction `mov dst src`のような命令
type Instruction struct {
	Op   Operation
	Args []Node
}

func (i Instruction) isNode() {}
func (i Instruction) String() string {
	var elms []string
	for _, elm := range i.Args {
		elms = append(elms, elm.String())
	}
	return fmt.Sprintf("Instruction{ Op: %s, Args: [ %s ] }", i.Op.String(), strings.Join(elms, ", "))
}

// Register `[sp+1]`の `sp`部分
type Register int

const (
	_ Register = iota
	PC
	SP
	BP
	HP

	R0
	R1
	R2
	R3
	R4
	R5
	R6
	R7
	R8
	R9
	R10

	ZF
)

func (r Register) isNode() {}
func (r Register) String() string {
	return []string{
		PC:  "pc",
		SP:  "sp",
		BP:  "bp",
		HP:  "hp",
		R0:  "r0",
		R1:  "r1",
		R2:  "r2",
		R3:  "r3",
		R4:  "r4",
		R5:  "r5",
		R6:  "r6",
		R7:  "r7",
		R8:  "r8",
		R9:  "r9",
		R10: "r10",
		ZF:  "zf",
	}[r]
}

// Offset `[sp+1]`のような相対位置
type Offset struct {
	Target Register
	Diff   int
}

func (o Offset) isNode() {}
func (o Offset) String() string {
	return fmt.Sprintf("[%s%+d]", o.Target.String(), o.Diff)
}

type Number int

func (n Number) isNode() {}
func (n Number) String() string {
	return strconv.Itoa(int(n))
}

type Character rune

func (c Character) isNode() {}
func (c Character) String() string {
	return strconv.QuoteRune(rune(c))
}

type Label struct {
	Define bool
	Name   string
}

func (l Label) isNode() {}
func (l Label) String() string {
	return l.Name
}

var curt *Token

func expect(kind TokenKind) (*Token, error) {
	if curt.Kind != kind {
		return nil, fmt.Errorf("want=%s, got=%s", kind.String(), curt.Kind.String())
	}
	v := *curt
	curt = curt.Next
	return &v, nil
}
func expectIdent(id string) error {
	if curt.Kind != Identifier {
		return fmt.Errorf("want=%s, got=%s", Identifier.String(), curt.Kind.String())
	}
	v := *curt
	if string(v.Raw) != id {
		return fmt.Errorf("want=%s, got=%s", id, string(v.Raw))
	}
	curt = curt.Next
	return nil
}
func consume(kind TokenKind) *Token {
	if curt.Kind != kind {
		return nil
	}
	v := *curt
	curt = curt.Next
	return &v
}
func consumeIdent(id string) *Token {
	if curt.Kind != Identifier {
		return nil
	}
	v := *curt
	if string(v.Raw) != id {
		return nil
	}
	curt = curt.Next
	return &v
}

func parseStackOffset() ([]Node, error) {
	// [
	if _, err := expect(Lcb); err != nil {
		return nil, err
	}

	// sp / bp
	id, err := expect(Identifier)
	if err != nil {
		return nil, err
	}
	var reg Register
	switch string(id.Raw) {
	case "sp":
		reg = SP
	case "bp":
		reg = BP
	default:
		return nil, fmt.Errorf("unsupported register: %s", string(id.Raw))
	}

	// +
	plus := consume(Add)
	// -
	minus := consume(Sub)
	if plus != nil && minus != nil {
		return nil, fmt.Errorf("syntax err")
	}

	// diff
	diff, err := expect(Integer)
	if err != nil {
		return nil, err
	}

	// ]
	if _, err := expect(Rcb); err != nil {
		return nil, err
	}

	v, err := diff.GetValueAsInteger()
	if err != nil {
		return nil, err
	}
	if minus != nil {
		return []Node{Offset{reg, -v}}, nil
	}
	return []Node{Offset{reg, v}}, nil
}

func parseLabel() ([]Node, error) {
	// id
	id, err := expect(Identifier)
	if err != nil {
		return nil, err
	}

	// :
	var define = false
	if t := consume(Colon); t != nil {
		define = true
	}

	return []Node{Label{define, string(id.Raw)}}, nil
}

func parseText() ([]Node, error) {
	var nodes []Node
loop:
	for {
		switch curt.Kind {
		case Eof:
			break loop
		case Comment:
			curt = curt.Next
		case Identifier:
			if op, yes := isOperation(string(curt.Raw)); yes {
				nodes = append(nodes, op)
				curt = curt.Next
				continue
			}
			if reg, yes := isRegister(string(curt.Raw)); yes {
				nodes = append(nodes, reg)
				curt = curt.Next
				continue
			}
			nds, err := parseLabel()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, nds...)
		case Integer:
			v, err := curt.GetValueAsInteger()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, Number(v))
			curt = curt.Next
		case Char:
			v, err := curt.GetValueAsRune()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, Character(v))
			curt = curt.Next
		case Lcb:
			nds, err := parseStackOffset()
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, nds...)
		default:
			return nil, fmt.Errorf("parse: unsupported token: %s", curt.Kind.String())
		}
	}
	return nodes, nil
}

type DataMode int

const (
	AUTO DataMode = iota
)

type ConstantData interface {
	isData()
	String() string
}

type ConstChar string

func (c ConstChar) isData() {}
func (c ConstChar) String() string {
	return string(c)
}

type ConstInt int

func (c ConstInt) isData() {}
func (c ConstInt) String() string {
	return strconv.Itoa(int(c))
}

type Constant struct {
	Name   string
	Mode   DataMode
	Values []ConstantData
}

type IR struct {
	Imports    []string
	Exports    []string
	Constants  map[string]Constant
	EntryPoint string
	Text       []Node
}

type ParseMode int

func parseImport() (string, error) {
	id, err := expect(Identifier)
	if err != nil {
		return "", err
	}
	return string(id.Raw), nil
}

func parseExport() (string, error) {
	id, err := expect(Identifier)
	if err != nil {
		return "", nil
	}
	return string(id.Raw), nil
}

func parseArray() ([]ConstantData, error) {
	// "hi" -> 'h','i'
	if str := consume(String); str != nil {
		var arr []ConstantData
		for _, r := range str.Raw {
			arr = append(arr, ConstChar(r))
		}
		return arr, nil
	}

	// arr
	var arr []ConstantData
	for curt.Kind != Eof {
		if i := consume(Integer); i != nil {
			i64, err := strconv.ParseInt(string(i.Raw), 10, 64)
			if err != nil {
				return nil, err
			}
			arr = append(arr, ConstInt(int(i64)))
		} else if c := consume(Char); c != nil {
			arr = append(arr, ConstChar(c.Raw[0]))
		}

		if comma := consume(Comma); comma == nil {
			break
		}
	}

	return arr, nil
}
func parseConstants() (map[string]Constant, error) {
	constants := map[string]Constant{}
	for curt.Kind != Eof {
		// msg, arr, ...
		id := consume(Identifier)
		if id == nil {
			break
		}

		// auto, ...
		var mode DataMode
		switch {
		case consumeIdent("auto") != nil:
			mode = AUTO
		default:
			return nil, fmt.Errorf("unsupported data mode: %s", curt.Kind.String())
		}

		// "hello", 10,10,10, 'h','i', ...
		arr, err := parseArray()
		if err != nil {
			return nil, err
		}

		constants[string(id.Raw)] = Constant{
			Name:   string(id.Raw),
			Mode:   mode,
			Values: arr,
		}
	}
	return constants, nil
}

func parseEntryPoint() (string, error) {
	if err := expectIdent("global"); err != nil {
		return "", err
	}
	id, err := expect(Identifier)
	if err != nil {
		return "", err
	}
	return string(id.Raw), nil
}

func Parse(token *Token) (*IR, error) {
	ir := IR{}
	ir.Imports = make([]string, 0)
	ir.Exports = make([]string, 0)
	ir.Constants = make(map[string]Constant)
	ir.EntryPoint = ""
	ir.Text = make([]Node, 0)

	curt = token
loop:
	for {
		switch curt.Kind {
		case Eof:
			break loop
		case Dot: // sections
			_, _ = expect(Dot)
			switch {
			case consumeIdent("import") != nil:
				import_, err := parseImport()
				if err != nil {
					return nil, err
				}
				ir.Imports = append(ir.Imports, import_)
			case consumeIdent("export") != nil:
				export, err := parseExport()
				if err != nil {
					return nil, err
				}
				ir.Exports = append(ir.Exports, export)
			case consumeIdent("section") != nil:
				_, err := expect(Dot)
				if err != nil {
					return nil, err
				}
				switch {
				case consumeIdent("data") != nil:
					_, err := expect(Colon)
					if err != nil {
						return nil, err
					}
					constants, err := parseConstants()
					if err != nil {
						return nil, err
					}
					ir.Constants = constants
				case consumeIdent("text") != nil:
					_, err := expect(Colon)
					if err != nil {
						return nil, err
					}
					entrypoint, err := parseEntryPoint()
					if err != nil {
						return nil, err
					}
					ir.EntryPoint = entrypoint
				default:
					return nil, fmt.Errorf("unsupported directive: %s", curt.Kind.String())
				}
			default:
				return nil, fmt.Errorf("unsupported directive: %s", curt.Kind.String())
			}
		default:
			program, err := parseText()
			if err != nil {
				return nil, err
			}
			ir.Text = program
		}
	}
	return &ir, nil
}

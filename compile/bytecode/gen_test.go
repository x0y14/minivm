package bytecode

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/x0y14/minivm/vm"
)

func TestGen(t *testing.T) {
	tests := []struct {
		name   string
		nodes  []Node
		expect []vm.Code
	}{
		{
			"nop",
			[]Node{
				Instruction{NOP, nil},
			},
			[]vm.Code{vm.NOP},
		},
		{
			"mov",
			[]Node{
				Instruction{MOV, []Node{Offset{BP, +1}, SP}},
				Instruction{MOV, []Node{Offset{SP, -2}, Number(1)}},
				Instruction{MOV, []Node{PC, Character('a')}},
			},
			[]vm.Code{
				vm.MOV, vm.BpOffset(1), vm.SP,
				vm.MOV, vm.SpOffset(-2), vm.Integer(1),
				vm.MOV, vm.PC, vm.Character('a'),
			},
		},
		{
			"fizzbuzz",
			[]Node{
				Instruction{ALLOC, []Node{Number(16)}},
				Instruction{POP, []Node{R10}},
				Instruction{MOV, []Node{R6, Number(1)}},

				Instruction{MOV, []Node{R5, Number(0)}},

				Instruction{MOV, []Node{R7, R6}},
				Instruction{MOV, []Node{R8, Number(3)}},

				Instruction{LT, []Node{R7, R8}},
				Instruction{JZ, []Node{Offset{PC, 7}}},
				Instruction{SUB, []Node{R7, R8}},
				Instruction{JMP, []Node{Offset{PC, -8}}},

				Instruction{EQ, []Node{R7, Number(0)}},
				Instruction{JZ, []Node{Offset{PC, 4}}},
				Instruction{JMP, []Node{Offset{PC, 30}}},

				// print_fizz:
				Instruction{STORE, []Node{R10, Character('F')}},
				Instruction{STORE, []Node{Number(1), Character('i')}},
				Instruction{STORE, []Node{Number(2), Character('z')}},
				Instruction{STORE, []Node{Number(3), Character('z')}},
				Instruction{MOV, []Node{R0, Number(1)}},
				Instruction{MOV, []Node{R1, Number(1)}},
				Instruction{MOV, []Node{R2, R10}},
				Instruction{MOV, []Node{R3, Number(4)}},
				Instruction{SYSCALL, nil},
				Instruction{MOV, []Node{R5, Number(1)}},

				// after_fizz:
				// --- i % 5 ---
				Instruction{MOV, []Node{R7, R6}},
				Instruction{MOV, []Node{R8, Number(5)}},
				Instruction{LT, []Node{R7, R8}},
				Instruction{JZ, []Node{Offset{PC, 7}}},
				Instruction{SUB, []Node{R7, R8}},
				Instruction{JMP, []Node{Offset{PC, -8}}},
				Instruction{EQ, []Node{R7, Number(0)}},
				Instruction{JZ, []Node{Offset{PC, 4}}},
				Instruction{JMP, []Node{Offset{PC, 30}}},

				// print_buzz:
				Instruction{STORE, []Node{R10, Character('B')}},
				Instruction{STORE, []Node{Number(1), Character('u')}},
				Instruction{STORE, []Node{Number(2), Character('z')}},
				Instruction{STORE, []Node{Number(3), Character('z')}},
				Instruction{MOV, []Node{R0, Number(1)}},
				Instruction{MOV, []Node{R1, Number(1)}},
				Instruction{MOV, []Node{R2, R10}},
				Instruction{MOV, []Node{R3, Number(4)}},
				Instruction{SYSCALL, nil},
				Instruction{MOV, []Node{R5, Number(1)}},

				// after_buzz:
				// printed == 0 なら数字を出力
				Instruction{EQ, []Node{R5, Number(0)}},
				Instruction{JZ, []Node{Offset{PC, 4}}},
				Instruction{JMP, []Node{Offset{PC, 78}}},

				// print_number:
				Instruction{LT, []Node{R6, Number(10)}},
				Instruction{JZ, []Node{Offset{PC, 51}}},

				// two_digit (10..15)
				Instruction{MOV, []Node{R3, Character('1')}},
				Instruction{STORE, []Node{R10, R3}},
				Instruction{MOV, []Node{R0, Number(1)}},
				Instruction{MOV, []Node{R1, Number(1)}},
				Instruction{MOV, []Node{R2, R10}},
				Instruction{MOV, []Node{R3, Number(1)}},
				Instruction{SYSCALL, nil},

				Instruction{MOV, []Node{R3, R6}},
				Instruction{SUB, []Node{R3, Number(10)}},
				Instruction{ADD, []Node{R3, Number('0')}},
				Instruction{STORE, []Node{Number(1), R3}},
				Instruction{MOV, []Node{R0, Number(1)}},
				Instruction{MOV, []Node{R1, Number(1)}},
				Instruction{MOV, []Node{R2, R10}},
				Instruction{ADD, []Node{R2, Number(1)}},
				Instruction{MOV, []Node{R3, Number(1)}},
				Instruction{SYSCALL, nil},
				Instruction{JMP, []Node{Offset{PC, 24}}},

				// one_digit:
				Instruction{MOV, []Node{R3, Number('0')}},
				Instruction{ADD, []Node{R3, R6}},
				Instruction{STORE, []Node{R10, R3}},
				Instruction{MOV, []Node{R0, Number(1)}},
				Instruction{MOV, []Node{R1, Number(1)}},
				Instruction{MOV, []Node{R2, R10}},
				Instruction{MOV, []Node{R3, Number(1)}},
				Instruction{SYSCALL, nil},

				// after_number:
				Instruction{STORE, []Node{R10, Character('\n')}},
				Instruction{MOV, []Node{R0, Number(1)}},
				Instruction{MOV, []Node{R1, Number(1)}},
				Instruction{MOV, []Node{R2, R10}},
				Instruction{MOV, []Node{R3, Number(1)}},
				Instruction{SYSCALL, nil},

				// i++
				Instruction{ADD, []Node{R6, Number(1)}},
				Instruction{LE, []Node{R6, Number(15)}},
				Instruction{JZ, []Node{Offset{PC, -210}}},
				Instruction{MOV, []Node{R0, Number(0)}},
				Instruction{SYSCALL, nil},
			},
			[]vm.Code{
				vm.ALLOC, vm.Integer(16),
				vm.POP, vm.R10, // R10 = bufBase(=0)
				vm.MOV, vm.R6, vm.Integer(1), // i = 1
				// loop_start:
				vm.MOV, vm.R5, vm.Integer(0), // printed = 0

				// --- i % 3 ---
				vm.MOV, vm.R7, vm.R6, // tmp = i
				vm.MOV, vm.R8, vm.Integer(3), // d = 3
				// mod3_loop:
				vm.LT, vm.R7, vm.R8, // if tmp < d then break
				vm.JZ, vm.PcOffset(7), // -> after_mod3
				vm.SUB, vm.R7, vm.R8, // tmp -= d
				vm.JMP, vm.PcOffset(-8), // -> mod3_loop
				// after_mod3:
				vm.EQ, vm.R7, vm.Integer(0), // tmp == 0 ?
				vm.JZ, vm.PcOffset(4), // -> print_fizz
				vm.JMP, vm.PcOffset(30), // -> after_fizz

				// print_fizz:
				vm.STORE, vm.R10, vm.Character('F'),
				vm.STORE, vm.Integer(1), vm.Character('i'),
				vm.STORE, vm.Integer(2), vm.Character('z'),
				vm.STORE, vm.Integer(3), vm.Character('z'),
				vm.MOV, vm.R0, vm.Integer(1), // SYS_WRITE
				vm.MOV, vm.R1, vm.Integer(1), // fd=1
				vm.MOV, vm.R2, vm.R10, // addr=buf
				vm.MOV, vm.R3, vm.Integer(4), // len=4
				vm.SYSCALL,
				vm.MOV, vm.R5, vm.Integer(1), // printed = 1

				// after_fizz:
				// --- i % 5 ---
				vm.MOV, vm.R7, vm.R6, // tmp = i
				vm.MOV, vm.R8, vm.Integer(5), // d = 5
				// mod5_loop:
				vm.LT, vm.R7, vm.R8,
				vm.JZ, vm.PcOffset(7), // -> after_mod5
				vm.SUB, vm.R7, vm.R8,
				vm.JMP, vm.PcOffset(-8), // -> mod5_loop
				// after_mod5:
				vm.EQ, vm.R7, vm.Integer(0),
				vm.JZ, vm.PcOffset(4), // -> print_buzz
				vm.JMP, vm.PcOffset(30), // -> after_buzz

				// print_buzz:
				vm.STORE, vm.R10, vm.Character('B'),
				vm.STORE, vm.Integer(1), vm.Character('u'),
				vm.STORE, vm.Integer(2), vm.Character('z'),
				vm.STORE, vm.Integer(3), vm.Character('z'),
				vm.MOV, vm.R0, vm.Integer(1), // SYS_WRITE
				vm.MOV, vm.R1, vm.Integer(1), // fd=1
				vm.MOV, vm.R2, vm.R10, // addr=buf
				vm.MOV, vm.R3, vm.Integer(4), // len=4
				vm.SYSCALL,
				vm.MOV, vm.R5, vm.Integer(1), // printed = 1

				// after_buzz:
				// printed == 0 なら数字を出力
				vm.EQ, vm.R5, vm.Integer(0),
				vm.JZ, vm.PcOffset(4), // -> print_number
				vm.JMP, vm.PcOffset(78), // -> after_number

				// print_number:
				vm.LT, vm.R6, vm.Integer(10),
				vm.JZ, vm.PcOffset(51), // -> one_digit

				// two_digit (10..15): '1' と下位桁で出力
				vm.MOV, vm.R3, vm.Character('1'),
				vm.STORE, vm.R10, vm.R3,
				vm.MOV, vm.R0, vm.Integer(1), // SYS_WRITE
				vm.MOV, vm.R1, vm.Integer(1), // fd=1
				vm.MOV, vm.R2, vm.R10, // addr=buf
				vm.MOV, vm.R3, vm.Integer(1), // len=1
				vm.SYSCALL,

				vm.MOV, vm.R3, vm.R6, // 下位桁 = i - 10 + '0'
				vm.SUB, vm.R3, vm.Integer(10),
				vm.ADD, vm.R3, vm.Integer('0'),
				vm.STORE, vm.Integer(1), vm.R3,
				vm.MOV, vm.R0, vm.Integer(1), // SYS_WRITE
				vm.MOV, vm.R1, vm.Integer(1), // fd=1
				vm.MOV, vm.R2, vm.R10, // addr=buf
				vm.ADD, vm.R2, vm.Integer(1), // addr+1
				vm.MOV, vm.R3, vm.Integer(1), // len=1
				vm.SYSCALL,
				vm.JMP, vm.PcOffset(24), // -> after_number

				// one_digit:
				vm.MOV, vm.R3, vm.Integer('0'),
				vm.ADD, vm.R3, vm.R6, // '0' + i
				vm.STORE, vm.R10, vm.R3,
				vm.MOV, vm.R0, vm.Integer(1), // SYS_WRITE
				vm.MOV, vm.R1, vm.Integer(1), // fd=1
				vm.MOV, vm.R2, vm.R10, // addr=buf
				vm.MOV, vm.R3, vm.Integer(1), // len=1
				vm.SYSCALL,

				// after_number:
				// 改行
				vm.STORE, vm.R10, vm.Character('\n'),
				vm.MOV, vm.R0, vm.Integer(1), // SYS_WRITE
				vm.MOV, vm.R1, vm.Integer(1), // fd=1
				vm.MOV, vm.R2, vm.R10, // addr=buf
				vm.MOV, vm.R3, vm.Integer(1), // len=1
				vm.SYSCALL,

				// i++
				vm.ADD, vm.R6, vm.Integer(1),
				vm.LE, vm.R6, vm.Integer(15), // i <= 15 ?
				vm.JZ, vm.PcOffset(-210), // -> loop_start
				vm.MOV, vm.R0, vm.Integer(0), // SYS_EXIT
				vm.SYSCALL,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes, err := Gen(tt.nodes)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.expect, codes); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

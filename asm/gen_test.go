package asm

import (
	"github.com/google/go-cmp/cmp"
	"testing"

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

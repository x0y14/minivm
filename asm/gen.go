package asm

import (
	"fmt"

	"github.com/x0y14/minivm/vm"
)

func convert(node Node) ([]vm.Code, error) {
	switch node := node.(type) {
	case Instruction:
		var result []vm.Code
		// op
		op, err := convert(node.Op)
		if err != nil {
			return nil, err
		}
		result = append(result, op...)
		// arguments
		for _, arg := range node.Args {
			a, err := convert(arg)
			if err != nil {
				return nil, err
			}
			result = append(result, a...)
		}
		return result, nil
	case Operation:
		op := []vm.Code{
			NOP:     vm.NOP,
			MOV:     vm.MOV,
			PUSH:    vm.PUSH,
			POP:     vm.POP,
			ALLOC:   vm.ALLOC,
			STORE:   vm.STORE,
			LOAD:    vm.LOAD,
			CALL:    vm.CALL,
			RET:     vm.RET,
			JMP:     vm.JMP,
			JZ:      vm.JZ,
			JNZ:     vm.JNZ,
			ADD:     vm.ADD,
			SUB:     vm.SUB,
			EQ:      vm.EQ,
			NE:      vm.NE,
			LT:      vm.LT,
			LE:      vm.LE,
			SYSCALL: vm.SYSCALL,
		}[node]
		return []vm.Code{op}, nil
	case Register:
		reg := []vm.Code{
			PC:  vm.PC,
			SP:  vm.SP,
			BP:  vm.BP,
			HP:  vm.HP,
			R0:  vm.R0,
			R1:  vm.R1,
			R2:  vm.R2,
			R3:  vm.R3,
			R4:  vm.R4,
			R5:  vm.R5,
			R6:  vm.R6,
			R7:  vm.R7,
			R8:  vm.R8,
			R9:  vm.R9,
			R10: vm.R10,
			ZF:  vm.ZF,
		}[node]
		return []vm.Code{reg}, nil
	case Offset:
		switch offset := node; offset.Target {
		case PC:
			return []vm.Code{vm.PcOffset(offset.Diff)}, nil
		case SP:
			return []vm.Code{vm.SpOffset(offset.Diff)}, nil
		case BP:
			return []vm.Code{vm.BpOffset(offset.Diff)}, nil
		default:
			return nil, fmt.Errorf("convert: unsupported offset: %s", offset.String())
		}
	case Number:
		return []vm.Code{vm.Integer(node)}, nil
	case Character:
		return []vm.Code{vm.Character(node)}, nil
	default:
		return nil, fmt.Errorf("convert: unsupported node: %s", node.String())
	}
}

func Gen(nodes []Node) ([]vm.Code, error) {
	var codes []vm.Code
	for _, node := range nodes {
		c, err := convert(node)
		if err != nil {
			return nil, err
		}
		codes = append(codes, c...)
	}
	return codes, nil
}

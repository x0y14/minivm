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
			NOP: vm.NOP,
			MOV: vm.MOV,
		}[node]
		return []vm.Code{op}, nil
	case Register:
		reg := []vm.Code{
			PC: vm.PC,
			SP: vm.SP,
			BP: vm.BP,
			HP: vm.HP,
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

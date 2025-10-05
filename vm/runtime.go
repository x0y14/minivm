package vm

import (
	"fmt"
	"io"
	"os"
)

type Config struct {
	StackSize int
	HeapSize  int
}

type registerSet struct {
	specials map[SpecialRegister]int
	generals map[GeneralPurposeRegister]Immediate
	flags    map[FlagRegister]bool
}

type Runtime struct {
	program   []Code
	registers registerSet
	stack     []Immediate
	heap      []Immediate
	halt      bool

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewRuntime(program []Code, config *Config) *Runtime {
	regs := registerSet{
		specials: map[SpecialRegister]int{
			PC: 0,
			BP: 0,
			SP: config.StackSize,
			HP: 0,
		},
		generals: map[GeneralPurposeRegister]Immediate{
			R0:  nil,
			R1:  nil,
			R2:  nil,
			R3:  nil,
			R4:  nil,
			R5:  nil,
			R6:  nil,
			R7:  nil,
			R8:  nil,
			R9:  nil,
			R10: nil,
		},
		flags: map[FlagRegister]bool{
			ZF: false,
		},
	}
	return &Runtime{
		program:   program,
		registers: regs,
		stack:     make([]Immediate, config.StackSize),
		heap:      make([]Immediate, config.HeapSize),
		halt:      false,

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// Register操作
func (r *Runtime) getSpecialReg(reg SpecialRegister) Integer {
	return Integer(r.registers.specials[reg])
}
func (r *Runtime) setSpecialReg(reg SpecialRegister, i Integer) {
	r.registers.specials[reg] = int(i)
}

func (r *Runtime) getGeneralReg(reg GeneralPurposeRegister) Immediate {
	return r.registers.generals[reg]
}
func (r *Runtime) setGeneralReg(reg GeneralPurposeRegister, imm Immediate) {
	r.registers.generals[reg] = imm
}

func (r *Runtime) getFlagReg(reg FlagRegister) Boolean {
	return Boolean(r.registers.flags[reg])
}
func (r *Runtime) setFlagReg(reg FlagRegister, boolean Boolean) {
	r.registers.flags[reg] = bool(boolean)
}

func (r *Runtime) getReg(reg Register) (Immediate, error) {
	switch reg.(type) {
	case SpecialRegister:
		return r.getSpecialReg(reg.(SpecialRegister)), nil
	case GeneralPurposeRegister:
		return r.getGeneralReg(reg.(GeneralPurposeRegister)), nil
	case FlagRegister:
		return r.getFlagReg(reg.(FlagRegister)), nil
	default:
		return nil, fmt.Errorf("getReg: unsupported register: %s", reg.String())
	}
}
func (r *Runtime) setReg(reg Register, imm Immediate) error {
	switch reg.(type) {
	case SpecialRegister:
		if _, ok := imm.(Integer); !ok {
			return fmt.Errorf("setReg: special register value must be Integer: %T", imm)
		}
		r.setSpecialReg(reg.(SpecialRegister), imm.(Integer))
		return nil
	case GeneralPurposeRegister:
		r.setGeneralReg(reg.(GeneralPurposeRegister), imm)
		return nil
	case FlagRegister:
		if _, ok := imm.(Boolean); !ok {
			return fmt.Errorf("setReg: flag register value must be Boolean: %T", imm)
		}
		r.setFlagReg(reg.(FlagRegister), imm.(Boolean))
		return nil
	default:
		return fmt.Errorf("setReg: unsupported register type: %T", reg)
	}
}

// Stack操作
func (r *Runtime) getStack(offset Offset) (Immediate, error) {
	switch offset.(type) {
	case BpOffset:
		curt := int(r.getSpecialReg(BP))
		diff := int(offset.(BpOffset))
		return r.stack[curt+diff], nil
	case SpOffset:
		curt := int(r.getSpecialReg(SP))
		diff := int(offset.(SpOffset))
		return r.stack[curt+diff], nil
	default:
		return nil, fmt.Errorf("getValueFromOffset: unsupported offset: %s", offset.String())
	}
}

func (r *Runtime) setStack(offset Offset, imm Immediate) error {
	switch offset.(type) {
	case BpOffset:
		curt := int(r.getSpecialReg(BP))
		diff := int(offset.(BpOffset))
		r.stack[curt+diff] = imm
		return nil
	case SpOffset:
		curt := int(r.getSpecialReg(SP))
		diff := int(offset.(SpOffset))
		r.stack[curt+diff] = imm
		return nil
	default:
		return fmt.Errorf("setStackValue: unsupported offset: %s", offset.String())
	}
}

func (r *Runtime) pushToStack(imm Immediate) error {
	r.setSpecialReg(SP, r.getSpecialReg(SP)-1)
	if r.getSpecialReg(SP) < 0 {
		return fmt.Errorf("pushToStack: stack overflow")
	}
	r.stack[r.getSpecialReg(SP)] = imm
	return nil
}
func (r *Runtime) popFromStack() (Immediate, error) {
	sp := r.getSpecialReg(SP)
	if len(r.stack) <= int(sp) {
		return nil, fmt.Errorf("popFromStack: stack underflow")
	}
	v := r.stack[sp]
	r.stack[sp] = nil
	r.setSpecialReg(SP, sp+1)
	return v, nil
}

// heap操作
func (r *Runtime) reserveHeap(size int) (Immediate, error) {
	if len(r.heap) <= int(r.getSpecialReg(HP))+size {
		return nil, fmt.Errorf("reserveHeap: out of memory")
	}
	baseAddr := r.getSpecialReg(HP)
	r.setSpecialReg(HP, r.getSpecialReg(HP)+Integer(size))
	return baseAddr, nil
}
func (r *Runtime) setHeap(heapAddr int, imm Immediate) error {
	if heapAddr < 0 || len(r.heap) <= heapAddr {
		return fmt.Errorf("setHeap: out of bounds")
	}
	r.heap[heapAddr] = imm
	return nil
}
func (r *Runtime) getHeap(heapAddr int) (Immediate, error) {
	if heapAddr < 0 || len(r.heap) <= heapAddr {
		return nil, fmt.Errorf("getHeap: out of bounds")
	}
	return r.heap[heapAddr], nil
}

func (r *Runtime) readHeapBytes(addr, length int) ([]byte, error) {
	if addr < 0 || length < 0 || addr+length > len(r.heap) {
		return nil, fmt.Errorf("readHeapBytes: out of bounds")
	}
	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		cell, err := r.getHeap(addr + i)
		if err != nil {
			return nil, err
		}
		if cell == nil {
			buf[i] = 0
			continue
		}
		buf[i] = byte(rune(cell.Value()))
	}
	return buf, nil
}
func (r *Runtime) writeHeapBytes(addr int, data []byte) error {
	if addr < 0 || addr+len(data) > len(r.heap) {
		return fmt.Errorf("writeHeapBytes: out of bounds")
	}
	for i, b := range data {
		if err := r.setHeap(addr+i, Character(int(b))); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runtime) relocate(op Opcode) {
	r.setSpecialReg(PC, Integer(int(r.getSpecialReg(PC))+op.NumOperands())+1)
}
func (r *Runtime) exec() error {
	switch code := r.program[r.getSpecialReg(PC)].(type) {
	case Opcode:
		switch code {
		case NOP:
			defer func() { r.relocate(code) }()
			return nil
		case MOV:
			defer func() { r.relocate(code) }()
			dst := r.program[r.getSpecialReg(PC)+1]
			src := r.program[r.getSpecialReg(PC)+2]

			var srcValue Immediate
			switch src.(type) {
			case Register:
				v, err := r.getReg(src.(Register))
				if err != nil {
					return err
				}
				srcValue = v
			case Offset:
				v, err := r.getStack(src.(Offset))
				if err != nil {
					return err
				}
				srcValue = v
			case Immediate:
				srcValue = src.(Immediate)
			default:
				return fmt.Errorf("mov: unsupported src: %s", src.String())
			}

			switch dst.(type) {
			case Register:
				return r.setReg(dst.(Register), srcValue)
			case Offset:
				return r.setStack(dst.(Offset), srcValue)
			default:
				return fmt.Errorf("mov: unsupported dst: %s", dst.String())
			}
		case PUSH:
			defer func() { r.relocate(code) }()
			src := r.program[r.getSpecialReg(PC)+1]

			var srcValue Immediate
			switch src.(type) {
			case Register:
				v, err := r.getReg(src.(Register))
				if err != nil {
					return err
				}
				srcValue = v
			case Offset:
				v, err := r.getStack(src.(Offset))
				if err != nil {
					return err
				}
				srcValue = v
			case Immediate:
				srcValue = src.(Immediate)
			default:
				return fmt.Errorf("push: unsupported src: %s", src.String())
			}
			return r.pushToStack(srcValue)
		case POP:
			defer func() { r.relocate(code) }()
			dst := r.program[r.getSpecialReg(PC)+1]
			switch dst.(type) {
			case Register:
				v, err := r.popFromStack()
				if err != nil {
					return err
				}
				return r.setReg(dst.(Register), v)
			default:
				return fmt.Errorf("pop: unsupported dst: %s", dst.String())
			}
		case ALLOC:
			defer func() { r.relocate(code) }()
			src := r.program[r.getSpecialReg(PC)+1]
			var size Immediate
			switch src.(type) {
			case Register:
				v, err := r.getReg(src.(Register))
				if err != nil {
					return err
				}
				size = v
			case Offset:
				v, err := r.getStack(src.(Offset))
				if err != nil {
					return err
				}
				size = v
			case Immediate:
				size = src.(Immediate)
			default:
				return fmt.Errorf("alloc: unsupported src: %s", src.String())
			}
			sizeValue := size.Value()
			baseAddr, err := r.reserveHeap(sizeValue)
			if err != nil {
				return err
			}
			return r.pushToStack(baseAddr)
		case STORE:
			defer func() { r.relocate(code) }()
			dst := r.program[r.getSpecialReg(PC)+1]
			src := r.program[r.getSpecialReg(PC)+2]

			var srcValue Immediate
			switch src.(type) {
			case Register:
				v, err := r.getReg(src.(Register))
				if err != nil {
					return err
				}
				srcValue = v
			case Offset:
				v, err := r.getStack(src.(Offset))
				if err != nil {
					return err
				}
				srcValue = v
			case Immediate:
				srcValue = src.(Immediate)
			default:
				return fmt.Errorf("store: unsupported src: %s", src.String())
			}

			switch dst.(type) {
			case Register:
				addr, err := r.getReg(dst.(Register))
				if err != nil {
					return err
				}
				if _, ok := addr.(Integer); !ok {
					return fmt.Errorf("store: unsupported dst: %s", dst.String())
				}
				return r.setHeap(int(addr.(Integer)), srcValue)
			case Offset:
				offset, err := r.getStack(dst.(Offset))
				if err != nil {
					return err
				}
				if _, ok := offset.(Integer); !ok {
					return fmt.Errorf("store: unsupported dst: %s", dst.String())
				}
				return r.setHeap(int(offset.(Integer)), srcValue)
			case Immediate:
				if _, ok := dst.(Integer); !ok {
					return fmt.Errorf("store: unsupported dst: %s", dst.String())
				}
				return r.setHeap(int(dst.(Integer)), srcValue)
			default:
				return fmt.Errorf("store: unsupported dst: %s", dst.String())
			}
		case LOAD:
			defer func() { r.relocate(code) }()
			dst := r.program[r.getSpecialReg(PC)+1]
			src := r.program[r.getSpecialReg(PC)+2]

			var srcValue Immediate
			switch src.(type) {
			case Register:
				v, err := r.getReg(src.(Register))
				if err != nil {
					return err
				}
				srcValue = v
			case Offset:
				v, err := r.getStack(src.(Offset))
				if err != nil {
					return err
				}
				srcValue = v
			case Immediate:
				srcValue = src.(Immediate)
			default:
				return fmt.Errorf("load: unsupported src: %s", src.String())
			}

			v, err := r.getHeap(srcValue.Value())
			if err != nil {
				return err
			}

			switch dst.(type) {
			case Register:
				return r.setReg(dst.(Register), v)
			case Offset:
				return r.setStack(dst.(Offset), v)
			default:
				return fmt.Errorf("load: unsupported dst: %s", dst.String())
			}
		case CALL:
			dst := r.program[r.getSpecialReg(PC)+1]
			if _, ok := dst.(PcOffset); !ok {
				return fmt.Errorf("call: unsupported dst: %s", dst.String())
			}
			// 現在のpcから計算
			base := int(r.getSpecialReg(PC))
			diff := int(dst.(PcOffset))
			dstAddr := Integer(base + diff)

			// relocate先と同じ
			retAddr := Integer(int(r.getSpecialReg(PC)) + code.NumOperands() + 1)
			if err := r.pushToStack(retAddr); err != nil {
				return err
			}
			r.setSpecialReg(PC, dstAddr)
			return nil
		case RET:
			dst, err := r.popFromStack()
			if err != nil {
				return err
			}
			switch dst.(type) {
			case Integer:
				r.setSpecialReg(PC, dst.(Integer))
				return nil
			default:
				return fmt.Errorf("ret: unsupported dst: %s", dst.String())
			}
		case JMP:
			dst := r.program[r.getSpecialReg(PC)+1]
			if _, ok := dst.(PcOffset); !ok {
				return fmt.Errorf("jmp: unsupported dst: %s", dst.String())
			}
			// 現在のpcから計算
			base := int(r.getSpecialReg(PC))
			diff := int(dst.(PcOffset))
			dstAddr := Integer(base + diff)
			r.setSpecialReg(PC, dstAddr)
			return nil
		case JZ:
			dst := r.program[r.getSpecialReg(PC)+1]
			if _, ok := dst.(PcOffset); !ok {
				return fmt.Errorf("je: unsupported dst: %s", dst.String())
			}
			// 現在のpcから計算
			base := int(r.getSpecialReg(PC))
			diff := int(dst.(PcOffset))
			dstAddr := Integer(base + diff)
			if r.getFlagReg(ZF) {
				r.setSpecialReg(PC, dstAddr)
				return nil
			}
			r.relocate(code)
			return nil
		case JNZ:
			dst := r.program[r.getSpecialReg(PC)+1]
			if _, ok := dst.(PcOffset); !ok {
				return fmt.Errorf("jne: unsupported dst: %s", dst.String())
			}
			// 現在のpcから計算
			base := int(r.getSpecialReg(PC))
			diff := int(dst.(PcOffset))
			dstAddr := Integer(base + diff)
			if !r.getFlagReg(ZF) {
				r.setSpecialReg(PC, dstAddr)
				return nil
			}
			r.relocate(code)
			return nil
		case ADD:
			defer func() { r.relocate(code) }()
			dst := r.program[r.getSpecialReg(PC)+1]
			src := r.program[r.getSpecialReg(PC)+2]

			var srcValue Immediate
			switch src.(type) {
			case Register:
				v, err := r.getReg(src.(Register))
				if err != nil {
					return err
				}
				srcValue = v
			case Offset:
				v, err := r.getStack(src.(Offset))
				if err != nil {
					return err
				}
				srcValue = v
			case Immediate:
				srcValue = src.(Immediate)
			default:
				return fmt.Errorf("add: unsupported src: %s", src.String())
			}

			switch dst.(type) {
			case Register:
				v, err := r.getReg(dst.(Register))
				if err != nil {
					return err
				}
				if !calculable(v, srcValue) {
					return fmt.Errorf("add: unsupported values: %T += %T", v, srcValue)
				}
				res := Integer(v.Value() + srcValue.Value())
				r.setFlagReg(ZF, res.Value() == 0)
				return r.setReg(dst.(Register), res)
			case Offset:
				v, err := r.getStack(dst.(Offset))
				if err != nil {
					return err
				}
				if !calculable(v, srcValue) {
					return fmt.Errorf("add: unsupported values: %T += %T", v, srcValue)
				}
				res := Integer(v.Value() + srcValue.Value())
				r.setFlagReg(ZF, res.Value() == 0)
				return r.setStack(dst.(Offset), res)
			default:
				return fmt.Errorf("add: unsupported dst: %s", dst.String())
			}
		case SUB:
			defer func() { r.relocate(code) }()
			dst := r.program[r.getSpecialReg(PC)+1]
			src := r.program[r.getSpecialReg(PC)+2]

			var srcValue Immediate
			switch src.(type) {
			case Register:
				v, err := r.getReg(src.(Register))
				if err != nil {
					return err
				}
				srcValue = v
			case Offset:
				v, err := r.getStack(src.(Offset))
				if err != nil {
					return err
				}
				srcValue = v
			case Immediate:
				srcValue = src.(Immediate)
			default:
				return fmt.Errorf("sub: unsupported src: %s", src.String())
			}

			switch dst.(type) {
			case Register:
				v, err := r.getReg(dst.(Register))
				if err != nil {
					return err
				}
				if !calculable(v, srcValue) {
					return fmt.Errorf("sub: unsupported values: %T -= %T", v, srcValue)
				}
				res := Integer(v.Value() - srcValue.Value())
				r.setFlagReg(ZF, res.Value() == 0)
				return r.setReg(dst.(Register), res)
			case Offset:
				v, err := r.getStack(dst.(Offset))
				if err != nil {
					return err
				}
				if !calculable(v, srcValue) {
					return fmt.Errorf("sub: unsupported values: %T -= %T", v, srcValue)
				}
				res := Integer(v.Value() - srcValue.Value())
				r.setFlagReg(ZF, res.Value() == 0)
				return r.setStack(dst.(Offset), res)
			default:
				return fmt.Errorf("sub: unsupported dst: %s", dst.String())
			}
		case EQ:
			defer func() { r.relocate(code) }()
			lhs := r.program[r.getSpecialReg(PC)+1]
			rhs := r.program[r.getSpecialReg(PC)+2]

			var lhsValue Immediate
			switch lhs.(type) {
			case Register:
				v, err := r.getReg(lhs.(Register))
				if err != nil {
					return err
				}
				lhsValue = v
			case Offset:
				v, err := r.getStack(lhs.(Offset))
				if err != nil {
					return err
				}
				lhsValue = v
			case Immediate:
				lhsValue = lhs.(Immediate)
			default:
				return fmt.Errorf("eq: unsupported lhs: %s", lhs.String())
			}

			var rhsValue Immediate
			switch rhs.(type) {
			case Register:
				v, err := r.getReg(rhs.(Register))
				if err != nil {
					return err
				}
				rhsValue = v
			case Offset:
				v, err := r.getStack(rhs.(Offset))
				if err != nil {
					return err
				}
				rhsValue = v
			case Immediate:
				rhsValue = rhs.(Immediate)
			default:
				return fmt.Errorf("eq: unsupported rhs: %s", rhs.String())
			}

			if lhsValue.Value() == rhsValue.Value() {
				r.setFlagReg(ZF, true)
				return nil
			}
			r.setFlagReg(ZF, false)
			return nil
		case NE:
			defer func() { r.relocate(code) }()
			lhs := r.program[r.getSpecialReg(PC)+1]
			rhs := r.program[r.getSpecialReg(PC)+2]

			var lhsValue Immediate
			switch lhs.(type) {
			case Register:
				v, err := r.getReg(lhs.(Register))
				if err != nil {
					return err
				}
				lhsValue = v
			case Offset:
				v, err := r.getStack(lhs.(Offset))
				if err != nil {
					return err
				}
				lhsValue = v
			case Immediate:
				lhsValue = lhs.(Immediate)
			default:
				return fmt.Errorf("ne: unsupported lhs: %s", lhs.String())
			}

			var rhsValue Immediate
			switch rhs.(type) {
			case Register:
				v, err := r.getReg(rhs.(Register))
				if err != nil {
					return err
				}
				rhsValue = v
			case Offset:
				v, err := r.getStack(rhs.(Offset))
				if err != nil {
					return err
				}
				rhsValue = v
			case Immediate:
				rhsValue = rhs.(Immediate)
			default:
				return fmt.Errorf("ne: unsupported rhs: %s", rhs.String())
			}

			if lhsValue.Value() != rhsValue.Value() {
				r.setFlagReg(ZF, true)
				return nil
			}
			r.setFlagReg(ZF, false)
			return nil
		case LT:
			defer func() { r.relocate(code) }()
			lhs := r.program[r.getSpecialReg(PC)+1]
			rhs := r.program[r.getSpecialReg(PC)+2]

			var lhsValue Immediate
			switch lhs.(type) {
			case Register:
				v, err := r.getReg(lhs.(Register))
				if err != nil {
					return err
				}
				lhsValue = v
			case Offset:
				v, err := r.getStack(lhs.(Offset))
				if err != nil {
					return err
				}
				lhsValue = v
			case Immediate:
				lhsValue = lhs.(Immediate)
			default:
				return fmt.Errorf("lt: unsupported lhs: %s", lhs.String())
			}

			var rhsValue Immediate
			switch rhs.(type) {
			case Register:
				v, err := r.getReg(rhs.(Register))
				if err != nil {
					return err
				}
				rhsValue = v
			case Offset:
				v, err := r.getStack(rhs.(Offset))
				if err != nil {
					return err
				}
				rhsValue = v
			case Immediate:
				rhsValue = rhs.(Immediate)
			default:
				return fmt.Errorf("lt: unsupported rhs: %s", rhs.String())
			}

			if lhsValue.Value() < rhsValue.Value() {
				r.setFlagReg(ZF, true)
				return nil
			}
			r.setFlagReg(ZF, false)
			return nil
		case LE:
			defer func() { r.relocate(code) }()
			lhs := r.program[r.getSpecialReg(PC)+1]
			rhs := r.program[r.getSpecialReg(PC)+2]

			var lhsValue Immediate
			switch lhs.(type) {
			case Register:
				v, err := r.getReg(lhs.(Register))
				if err != nil {
					return err
				}
				lhsValue = v
			case Offset:
				v, err := r.getStack(lhs.(Offset))
				if err != nil {
					return err
				}
				lhsValue = v
			case Immediate:
				lhsValue = lhs.(Immediate)
			default:
				return fmt.Errorf("le: unsupported lhs: %s", lhs.String())
			}

			var rhsValue Immediate
			switch rhs.(type) {
			case Register:
				v, err := r.getReg(rhs.(Register))
				if err != nil {
					return err
				}
				rhsValue = v
			case Offset:
				v, err := r.getStack(rhs.(Offset))
				if err != nil {
					return err
				}
				rhsValue = v
			case Immediate:
				rhsValue = rhs.(Immediate)
			default:
				return fmt.Errorf("le: unsupported rhs: %s", rhs.String())
			}

			if lhsValue.Value() <= rhsValue.Value() {
				r.setFlagReg(ZF, true)
				return nil
			}
			r.setFlagReg(ZF, false)
			return nil
		case SYSCALL:
			defer func() { r.relocate(code) }()
			no, err := r.getReg(R0)
			if err != nil {
				return err
			}
			switch no.Value() {
			case SYS_EXIT:
				r.halt = true
				return nil
			case SYS_WRITE:
				fd := r.getGeneralReg(R1).Value()
				addr := r.getGeneralReg(R2).Value()
				length := r.getGeneralReg(R3).Value()
				data, err := r.readHeapBytes(addr, length)
				if err != nil {
					return err
				}

				var w io.Writer
				switch fd {
				case 1:
					w = r.stdout
				case 2:
					w = r.stderr
				default:
					return fmt.Errorf("sys_write: unsupported fd: %d", fd)
				}
				wrote, err := w.Write(data)
				if err != nil {
					return err
				}
				return r.setReg(R0, Integer(wrote))
			case SYS_READ:
				fd := r.getGeneralReg(R1).Value()
				addr := r.getGeneralReg(R2).Value()
				length := r.getGeneralReg(R3).Value()

				var rd io.Reader
				switch fd {
				case 0:
					rd = r.stdin
				default:
					return fmt.Errorf("sys_read: unsupported fd: %d", fd)
				}
				buf := make([]byte, length)
				got, err := rd.Read(buf)
				if err != nil {
					return err
				}
				if got > 0 {
					if err := r.writeHeapBytes(addr, buf[:got]); err != nil {
						return err
					}
				}
				return r.setReg(R0, Integer(got))
			default:
				return fmt.Errorf("syscall: unsupported syscallNo: %d", no.Value())
			}
		default:
			return fmt.Errorf("exec: unimplemented opcode: %s", code.String())
		}
	default:
		return fmt.Errorf("exec: unimplemented code: %s", code.String())
	}
}
func (r *Runtime) Run() error {
	for {
		if r.halt {
			return nil
		}
		switch code := r.program[r.getSpecialReg(PC)]; code.(type) {
		case Opcode:
			if err := r.exec(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported code: %s", code.String())
		}
	}
}

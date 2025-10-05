package minivm

import (
	"testing"
)

func TestNOP(t *testing.T) {
	program := []Code{
		NOP,
		MOV, R1, Integer(0),
		SYSCALL,
	}
	config := &Config{
		StackSize: 100,
		HeapSize:  100,
	}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v, wantErr %v", err, nil)
	}

	// R1は0のまま
	if runtime.registers.generals[R1] != Integer(0) {
		t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], Integer(0))
	}
}

func TestMOV(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantErr bool
		wantR0  Immediate
		wantR1  Immediate
	}{
		{
			name: "mov register to register",
			program: []Code{
				MOV, R0, Integer(42),
				MOV, R1, R0,
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantErr: false,
			wantR0:  Integer(42),
			wantR1:  Integer(0),
		},
		{
			name: "mov immediate to register",
			program: []Code{
				MOV, R0, Integer(100),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantErr: false,
			wantR0:  Integer(100),
			wantR1:  Integer(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StackSize: 100, HeapSize: 100}
			runtime := NewRuntime(tt.program, config)
			err := runtime.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if runtime.registers.generals[R0] != tt.wantR0 {
				t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], tt.wantR0)
			}
			if runtime.registers.generals[R1] != tt.wantR1 {
				t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], tt.wantR1)
			}
		})
	}
}

func TestPUSHAndPOP(t *testing.T) {
	program := []Code{
		PUSH, Integer(42),
		POP, R0,
		MOV, R1, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0に42が格納されているはず
	if runtime.registers.generals[R0] != Integer(42) {
		t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], Integer(42))
	}
}

func TestADD(t *testing.T) {
	program := []Code{
		MOV, R0, Integer(10),
		ADD, R0, Integer(5),
		MOV, R1, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0は10+5=15になっているはず
	if runtime.registers.generals[R0] != Integer(15) {
		t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], Integer(15))
	}
}

func TestSUB(t *testing.T) {
	program := []Code{
		MOV, R0, Integer(10),
		SUB, R0, Integer(3),
		MOV, R1, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0は10-3=7になっているはず
	if runtime.registers.generals[R0] != Integer(7) {
		t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], Integer(7))
	}
}

func TestEQAndNE(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantZF  bool
	}{
		{
			name: "eq: equal values",
			program: []Code{
				MOV, R0, Integer(10),
				EQ, R0, Integer(10),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
		{
			name: "eq: not equal values",
			program: []Code{
				MOV, R0, Integer(10),
				EQ, R0, Integer(5),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: false,
		},
		{
			name: "ne: equal values",
			program: []Code{
				MOV, R0, Integer(10),
				NE, R0, Integer(10),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: false,
		},
		{
			name: "ne: not equal values",
			program: []Code{
				MOV, R0, Integer(10),
				NE, R0, Integer(5),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StackSize: 100, HeapSize: 100}
			runtime := NewRuntime(tt.program, config)
			if err := runtime.Run(); err != nil {
				t.Errorf("Run() error = %v", err)
			}
			if runtime.registers.flags[ZF] != tt.wantZF {
				t.Errorf("ZF = %v, want %v", runtime.registers.flags[ZF], tt.wantZF)
			}
		})
	}
}

func TestJMP(t *testing.T) {
	program := []Code{
		JMP, PcOffset(4), // skip next instruction
		MOV, R0, Integer(999), // skipped
		MOV, R1, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0は初期値のnilのまま（999が代入されていないことを確認）
	if runtime.registers.generals[R0] != nil {
		t.Errorf("R0 = %v, want nil (instruction should be skipped)", runtime.registers.generals[R0])
	}
}

func TestJE(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantR0  Immediate
	}{
		{
			name: "jump when equal",
			program: []Code{
				EQ, Integer(1), Integer(1),
				JZ, PcOffset(4), // jump if equal
				MOV, R0, Integer(999), // skipped
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantR0: nil, // skipped
		},
		{
			name: "no jump when not equal",
			program: []Code{
				EQ, Integer(1), Integer(2),
				JZ, PcOffset(2), // no jump
				MOV, R0, Integer(999), // executed
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantR0: Integer(999), // executed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StackSize: 100, HeapSize: 100}
			runtime := NewRuntime(tt.program, config)
			if err := runtime.Run(); err != nil {
				t.Errorf("Run() error = %v", err)
			}
			if runtime.registers.generals[R0] != tt.wantR0 {
				t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], tt.wantR0)
			}
		})
	}
}

func TestJNE(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantR0  Immediate
	}{
		{
			name: "jump when not zero",
			program: []Code{
				EQ, Integer(1), Integer(2),
				JNZ, PcOffset(4), // jump if not zero
				MOV, R0, Integer(999), // skipped
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantR0: nil, // skipped
		},
		{
			name: "no jump when zero",
			program: []Code{
				EQ, Integer(1), Integer(1),
				JNZ, PcOffset(4), // no jump
				MOV, R0, Integer(999), // executed
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantR0: Integer(999),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StackSize: 100, HeapSize: 100}
			runtime := NewRuntime(tt.program, config)
			if err := runtime.Run(); err != nil {
				t.Errorf("Run() error = %v", err)
			}
			if runtime.registers.generals[R0] != tt.wantR0 {
				t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], tt.wantR0)
			}
		})
	}
}

func TestCALLAndRET(t *testing.T) {
	program := []Code{
		CALL, PcOffset(5), // call function
		MOV, R1, Integer(0),
		SYSCALL,
		// function starts here
		MOV, R0, Integer(42),
		RET,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}
	if runtime.registers.generals[R0] != Integer(42) {
		t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], Integer(42))
	}
}

func TestALLOC(t *testing.T) {
	program := []Code{
		ALLOC, Integer(10),
		POP, R0,
		MOV, R1, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0にはヒープのベースアドレス(0)が格納されているはず
	if runtime.registers.generals[R0] != Integer(0) {
		t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], Integer(0))
	}

	// ヒープポインタが10進んでいるはず
	if runtime.registers.specials[HP] != 10 {
		t.Errorf("HP = %v, want %v", runtime.registers.specials[HP], 10)
	}
}

func TestSTOREAndLOAD(t *testing.T) {
	program := []Code{
		ALLOC, Integer(1), // size = Integer(1)
		POP, R0, // R0 = baseAddr
		STORE, R0, Integer(42),
		LOAD, R2, R0,
		MOV, R1, Integer(0), // Exit
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R2には42がロードされているはず
	if runtime.registers.generals[R2] != Integer(42) {
		t.Errorf("R2 = %v, want %v", runtime.registers.generals[R2], Integer(42))
	}
}

func TestLT(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantZF  bool
	}{
		{
			name: "5 < 10 is true",
			program: []Code{
				LT, Integer(5), Integer(10),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
		{
			name: "10 < 5 is false",
			program: []Code{
				LT, Integer(10), Integer(5),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StackSize: 100, HeapSize: 100}
			runtime := NewRuntime(tt.program, config)
			if err := runtime.Run(); err != nil {
				t.Errorf("Run() error = %v", err)
			}
			if runtime.registers.flags[ZF] != tt.wantZF {
				t.Errorf("ZF = %v, want %v", runtime.registers.flags[ZF], tt.wantZF)
			}
		})
	}
}

func TestLE(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantZF  bool
	}{
		{
			name: "10 <= 10 is true",
			program: []Code{
				LE, Integer(10), Integer(10),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
		{
			name: "5 <= 10 is true",
			program: []Code{
				LE, Integer(5), Integer(10),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
		{
			name: "10 <= 5 is false",
			program: []Code{
				LE, Integer(10), Integer(5),
				MOV, R1, Integer(0),
				SYSCALL,
			},
			wantZF: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StackSize: 100, HeapSize: 100}
			runtime := NewRuntime(tt.program, config)
			if err := runtime.Run(); err != nil {
				t.Errorf("Run() error = %v", err)
			}
			if runtime.registers.flags[ZF] != tt.wantZF {
				t.Errorf("ZF = %v, want %v", runtime.registers.flags[ZF], tt.wantZF)
			}
		})
	}
}

func TestStackOffset(t *testing.T) {
	program := []Code{
		PUSH, Integer(42),
		MOV, R0, SpOffset(0),
		POP, R1,
		MOV, R1, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0にはスタックから読み取った42が格納されているはず
	if runtime.registers.generals[R0] != Integer(42) {
		t.Errorf("R0 = %v, want %v", runtime.registers.generals[R0], Integer(42))
	}
}

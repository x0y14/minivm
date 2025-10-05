package vm

import (
	"bytes"
	"testing"
)

func TestNOP(t *testing.T) {
	program := []Code{
		NOP,
		MOV, R0, Integer(0),
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
}

func TestMOV(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantErr bool
		wantR1  Immediate
		wantR2  Immediate
	}{
		{
			name: "mov register to register",
			program: []Code{
				MOV, R1, Integer(42),
				MOV, R2, R1,
				MOV, R2, Integer(0),
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantErr: false,
			wantR1:  Integer(42),
			wantR2:  Integer(0),
		},
		{
			name: "mov immediate to register",
			program: []Code{
				MOV, R1, Integer(100),
				MOV, R2, Integer(0),
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantErr: false,
			wantR1:  Integer(100),
			wantR2:  Integer(0),
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
			if runtime.registers.generals[R1] != tt.wantR1 {
				t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], tt.wantR1)
			}
			if runtime.registers.generals[R2] != tt.wantR2 {
				t.Errorf("R2 = %v, want %v", runtime.registers.generals[R2], tt.wantR2)
			}
		})
	}
}

func TestPUSHAndPOP(t *testing.T) {
	program := []Code{
		PUSH, Integer(42),
		POP, R1,
		MOV, R0, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0に42が格納されているはず
	if runtime.registers.generals[R1] != Integer(42) {
		t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], Integer(42))
	}
}

func TestADD(t *testing.T) {
	program := []Code{
		MOV, R1, Integer(10),
		ADD, R1, Integer(5),
		MOV, R0, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0は10+5=15になっているはず
	if runtime.registers.generals[R1] != Integer(15) {
		t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], Integer(15))
	}
}

func TestSUB(t *testing.T) {
	program := []Code{
		MOV, R1, Integer(10),
		SUB, R1, Integer(3),
		MOV, R0, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0は10-3=7になっているはず
	if runtime.registers.generals[R1] != Integer(7) {
		t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], Integer(7))
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
				MOV, R1, Integer(10),
				EQ, R1, Integer(10),
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
		{
			name: "eq: not equal values",
			program: []Code{
				MOV, R1, Integer(10),
				EQ, R1, Integer(5),
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantZF: false,
		},
		{
			name: "ne: equal values",
			program: []Code{
				MOV, R1, Integer(10),
				NE, R1, Integer(10),
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantZF: false,
		},
		{
			name: "ne: not equal values",
			program: []Code{
				MOV, R1, Integer(10),
				NE, R1, Integer(5),
				MOV, R0, Integer(0),
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
		JMP, PcOffset(5), // skip next instruction
		MOV, R1, Integer(999), // skipped
		MOV, R0, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0は初期値のnilのまま（999が代入されていないことを確認）
	if runtime.registers.generals[R1] != nil {
		t.Errorf("R1 = %v, want nil (instruction should be skipped)", runtime.registers.generals[R1])
	}
}

func TestJE(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantR1  Immediate
	}{
		{
			name: "jump when equal",
			program: []Code{
				EQ, Integer(1), Integer(1),
				JZ, PcOffset(5), // jump if equal
				MOV, R1, Integer(999), // skipped
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantR1: nil, // skipped
		},
		{
			name: "no jump when not equal",
			program: []Code{
				EQ, Integer(1), Integer(2),
				JZ, PcOffset(2), // no jump
				MOV, R1, Integer(999), // executed
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantR1: Integer(999), // executed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StackSize: 100, HeapSize: 100}
			runtime := NewRuntime(tt.program, config)
			if err := runtime.Run(); err != nil {
				t.Errorf("Run() error = %v", err)
			}
			if runtime.registers.generals[R1] != tt.wantR1 {
				t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], tt.wantR1)
			}
		})
	}
}

func TestJNE(t *testing.T) {
	tests := []struct {
		name    string
		program []Code
		wantR1  Immediate
	}{
		{
			name: "jump when not zero",
			program: []Code{
				EQ, Integer(1), Integer(2),
				JNZ, PcOffset(5), // jump if not zero
				MOV, R1, Integer(999), // skipped
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantR1: nil, // skipped
		},
		{
			name: "no jump when zero",
			program: []Code{
				EQ, Integer(1), Integer(1),
				JNZ, PcOffset(4), // no jump
				MOV, R1, Integer(999), // executed
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantR1: Integer(999),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{StackSize: 100, HeapSize: 100}
			runtime := NewRuntime(tt.program, config)
			if err := runtime.Run(); err != nil {
				t.Errorf("Run() error = %v", err)
			}
			if runtime.registers.generals[R1] != tt.wantR1 {
				t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], tt.wantR1)
			}
		})
	}
}

func TestCALLAndRET(t *testing.T) {
	program := []Code{
		CALL, PcOffset(6), // call function
		MOV, R0, Integer(0),
		SYSCALL,
		// function starts here
		MOV, R1, Integer(42),
		RET,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}
	if runtime.registers.generals[R1] != Integer(42) {
		t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], Integer(42))
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
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
		{
			name: "10 < 5 is false",
			program: []Code{
				LT, Integer(10), Integer(5),
				MOV, R0, Integer(0),
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
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
		{
			name: "5 <= 10 is true",
			program: []Code{
				LE, Integer(5), Integer(10),
				MOV, R0, Integer(0),
				SYSCALL,
			},
			wantZF: true,
		},
		{
			name: "10 <= 5 is false",
			program: []Code{
				LE, Integer(10), Integer(5),
				MOV, R0, Integer(0),
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
		MOV, R1, SpOffset(0),
		POP, R1,
		MOV, R0, Integer(0),
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	// R0にはスタックから読み取った42が格納されているはず
	if runtime.registers.generals[R1] != Integer(42) {
		t.Errorf("R1 = %v, want %v", runtime.registers.generals[R1], Integer(42))
	}
}

func TestSyscallWrite(t *testing.T) {
	// "hi" をヒープに書き込み、SYS_WRITEでstdoutに出力
	program := []Code{
		ALLOC, Integer(2), // ヒープ2バイト確保
		POP, R2, // R2 = baseAddr
		STORE, R2, Character('h'),
		STORE, Integer(1), Character('i'),
		MOV, R1, Integer(1), // fd=1(stdout)
		MOV, R3, Integer(2), // length=2
		MOV, R0, Integer(1), // syscall番号: SYS_WRITE
		SYSCALL,
		MOV, R0, Integer(0), // exit
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	var buf bytes.Buffer
	runtime.stdout = &buf

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}
	if buf.String() != "hi" {
		t.Errorf("stdout = %q, want %q", buf.String(), "hi")
	}
}

func TestSyscallRead(t *testing.T) {
	// stdinから2バイト読み込んでヒープに格納
	program := []Code{
		ALLOC, Integer(2), // ヒープ2バイト確保
		POP, R2, // R2 = baseAddr
		MOV, R1, Integer(0), // fd=0(stdin)
		MOV, R3, Integer(2), // length=2
		MOV, R0, Integer(2), // syscall番号: SYS_READ
		SYSCALL,
		LOAD, R4, R2, // R4 = heap[R2]
		LOAD, R5, Integer(1), // R5 = heap[R2+1]
		MOV, R0, Integer(0), // exit
		SYSCALL,
	}
	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	runtime.stdin = bytes.NewBufferString("hi")

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}
	if runtime.registers.generals[R4] != Character('h') {
		t.Errorf("R4 = %v, want %v", runtime.registers.generals[R4], Character('h'))
	}
	if runtime.registers.generals[R5] != Character('i') {
		t.Errorf("R5 = %v, want %v", runtime.registers.generals[R5], Character('i'))
	}
}

func TestFizzBuzzSimple(t *testing.T) {
	// まず"1\n"だけ出力するテスト
	program := []Code{
		ALLOC, Integer(2),
		POP, R10,
		MOV, R3, Integer('1'),
		STORE, R10, R3,
		MOV, R0, Integer(1), // SYS_WRITE
		MOV, R1, Integer(1), // fd=1
		MOV, R2, R10, // buffer
		MOV, R3, Integer(1), // length=1
		SYSCALL,
		// newline
		STORE, R10, Character('\n'),
		MOV, R0, Integer(1),
		MOV, R1, Integer(1),
		MOV, R2, R10,
		MOV, R3, Integer(1),
		SYSCALL,
		MOV, R0, Integer(0),
		SYSCALL,
	}

	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	var buf bytes.Buffer
	runtime.stdout = &buf

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	if buf.String() != "1\n" {
		t.Errorf("output = %q, want %q", buf.String(), "1\n")
	}
}

func TestModulo(t *testing.T) {
	// 5 % 3 = 2 を計算
	program := []Code{
		MOV, R3, Integer(5), // PC=0
		MOV, R4, Integer(3), // PC=2
		// mod loop (PC=4)
		LT, R3, R4, // PC=4: if R3 < 3, ZF=true
		JZ, PcOffset(7), // PC=6: if ZF=true, jump to 6+2+4=12
		SUB, R3, R4, // PC=8: R3 -= 3
		JMP, PcOffset(-8), // PC=10: jump to 10+2-8=4 (back to LT)
		// R3 should be 2 (PC=12)
		MOV, R1, R3, // PC=12
		MOV, R0, Integer(0), // PC=14
		SYSCALL, // PC=16
	}

	config := &Config{StackSize: 100, HeapSize: 100}
	runtime := NewRuntime(program, config)

	if err := runtime.Run(); err != nil {
		t.Errorf("Run() error = %v", err)
	}

	if runtime.registers.generals[R1] != Integer(2) {
		t.Errorf("R1 = %v, want %v (5 %% 3 = 2)", runtime.registers.generals[R1], Integer(2))
	}
}

func TestFizzBuzz(t *testing.T) {
	program := []Code{
		// ヒープに作業バッファ確保（先頭アドレスは 0）
		ALLOC, Integer(16),
		POP, R10, // R10 = bufBase(=0)
		MOV, R6, Integer(1), // i = 1
		// loop_start:
		MOV, R5, Integer(0), // printed = 0

		// --- i % 3 ---
		MOV, R7, R6, // tmp = i
		MOV, R8, Integer(3), // d = 3
		// mod3_loop:
		LT, R7, R8, // if tmp < d then break
		JZ, PcOffset(7), // -> after_mod3
		SUB, R7, R8, // tmp -= d
		JMP, PcOffset(-8), // -> mod3_loop
		// after_mod3:
		EQ, R7, Integer(0), // tmp == 0 ?
		JZ, PcOffset(4), // -> print_fizz
		JMP, PcOffset(30), // -> after_fizz

		// print_fizz:
		STORE, R10, Character('F'),
		STORE, Integer(1), Character('i'),
		STORE, Integer(2), Character('z'),
		STORE, Integer(3), Character('z'),
		MOV, R0, Integer(1), // SYS_WRITE
		MOV, R1, Integer(1), // fd=1
		MOV, R2, R10, // addr=buf
		MOV, R3, Integer(4), // len=4
		SYSCALL,
		MOV, R5, Integer(1), // printed = 1

		// after_fizz:
		// --- i % 5 ---
		MOV, R7, R6, // tmp = i
		MOV, R8, Integer(5), // d = 5
		// mod5_loop:
		LT, R7, R8,
		JZ, PcOffset(7), // -> after_mod5
		SUB, R7, R8,
		JMP, PcOffset(-8), // -> mod5_loop
		// after_mod5:
		EQ, R7, Integer(0),
		JZ, PcOffset(4), // -> print_buzz
		JMP, PcOffset(30), // -> after_buzz

		// print_buzz:
		STORE, R10, Character('B'),
		STORE, Integer(1), Character('u'),
		STORE, Integer(2), Character('z'),
		STORE, Integer(3), Character('z'),
		MOV, R0, Integer(1), // SYS_WRITE
		MOV, R1, Integer(1), // fd=1
		MOV, R2, R10, // addr=buf
		MOV, R3, Integer(4), // len=4
		SYSCALL,
		MOV, R5, Integer(1), // printed = 1

		// after_buzz:
		// printed == 0 なら数字を出力
		EQ, R5, Integer(0),
		JZ, PcOffset(4), // -> print_number
		JMP, PcOffset(78), // -> after_number

		// print_number:
		LT, R6, Integer(10),
		JZ, PcOffset(51), // -> one_digit

		// two_digit (10..15): '1' と下位桁で出力
		MOV, R3, Character('1'),
		STORE, R10, R3,
		MOV, R0, Integer(1), // SYS_WRITE
		MOV, R1, Integer(1), // fd=1
		MOV, R2, R10, // addr=buf
		MOV, R3, Integer(1), // len=1
		SYSCALL,

		MOV, R3, R6, // 下位桁 = i - 10 + '0'
		SUB, R3, Integer(10),
		ADD, R3, Integer('0'),
		STORE, Integer(1), R3,
		MOV, R0, Integer(1), // SYS_WRITE
		MOV, R1, Integer(1), // fd=1
		MOV, R2, R10, // addr=buf
		ADD, R2, Integer(1), // addr+1
		MOV, R3, Integer(1), // len=1
		SYSCALL,
		JMP, PcOffset(24), // -> after_number

		// one_digit:
		MOV, R3, Integer('0'),
		ADD, R3, R6, // '0' + i
		STORE, R10, R3,
		MOV, R0, Integer(1), // SYS_WRITE
		MOV, R1, Integer(1), // fd=1
		MOV, R2, R10, // addr=buf
		MOV, R3, Integer(1), // len=1
		SYSCALL,

		// after_number:
		// 改行
		STORE, R10, Character('\n'),
		MOV, R0, Integer(1), // SYS_WRITE
		MOV, R1, Integer(1), // fd=1
		MOV, R2, R10, // addr=buf
		MOV, R3, Integer(1), // len=1
		SYSCALL,

		// i++
		ADD, R6, Integer(1),
		LE, R6, Integer(15), // i <= 15 ?
		JZ, PcOffset(-210), // -> loop_start
		MOV, R0, Integer(0), // SYS_EXIT
		SYSCALL,
	}

	config := &Config{StackSize: 1024, HeapSize: 1024}
	rt := NewRuntime(program, config)

	var buf bytes.Buffer
	rt.stdout = &buf

	if err := rt.Run(); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := "1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz\n"
	if buf.String() != want {
		t.Errorf("output = %q, want %q", buf.String(), want)
	}
}

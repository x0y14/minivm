# miniVM

実験用に作った小さなVMです

スタックとヒープがあります

## Usage
show help
```shell
$ go run ./cmd/minivm/main.go help
```

run fizzbuzz
```shell
# bytecode
$ go run ./cmd/minivm/main.go run ./examples/bytecode/fizzbuzz.mbyt
# ir
$ go run  ./cmd/minivm/main.go run --link ./examples/ir/fizzbuzz.mir
```

brainf*ck
```shell
# 末尾に!をつけてください
$ go run ./cmd/minivm/main.go run -stack 32768 -heap 131072 ./examples/bytecode/brainfuck.mbyt
++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
++++++++++++++++++++++++++++++++++++++++.---.+++++++..+++.++++++
++.--------.+++.------.--------.!
> helloworld
```

## ファイル種別

### *.mir
ある程度人間が読める中間表現。ラベルとかがそのまま残っていて、リンク可能

### *.mbyt
なんちゃってバイトコード。vmはこれをインタプリタで逐次実行します

## ABI
### レジスタの種類
| Register名      | 用途              |
|----------------|-----------------|
| PC, SP, BP, HP | 特殊レジスタ          |
| ZF             | フラグレジスタ         |
| R0             | 戻り値 / システムコール番号 |
| R1 ~ R6        | 関数の引数           |
| R7 ~ R9        | 汎用              |
| R10            | 一時利用            |

### レジスタ保存規約
- caller-saved(呼び出し側が保存/復元する、呼び出された側はこれらを破壊して良い)
  - R0~R6, R10, ZF
- callee-saved(呼び出された側が保存/復元する)
  - R7~R9, BP, SP

### 引数と戻り値
- 引数: 最初の6個はR1~R6に。7以降はスタックに。([BP+2]を含むこれ以降に。)
- 戻り値: R0, 複数未対応

> [!NOTE]
> R0の扱いについて
> R0に終了コードを入れつつ、syscallの引数とするようなイメージ
> ```
> ; _start
> CALL  main
> MOV   R1, R0          ; R1 = status (mainの戻り値)
> MOV   R0, SYS_EXIT    ; R0 = syscall番号
> SYSCALL               ; プロセス終了（戻らない想定）
> ```

### システムコール
| 命令番号 | 命令名   | 引数                           | 用途             |
|------|-------|------------------------------|----------------|
| 0    | exit  | R1: status                   | プログラムを終了する際に使用 |
| 1    | write | R1: fd, R2: addr, R3: length | バッファ書き込み       |
| 2    | read  | R1: fd, R2: addr, R3: length | バッファ読み込み       |

## Linkについて

`_start`はエントリーポイントなので使用しないでください

`__`から始まるラベルはローカルラベルとします
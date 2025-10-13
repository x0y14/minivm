# Fizzbuzz

Python で 以下を実行した場合と同じ出力を得られます

```python
for i in range(1, 101):
    if i % 3 == 0 and i % 5 == 0:
        print('FizzBuzz')
    elif i % 3 == 0:
        print('Fizz')
    elif i % 5 == 0:
        print('Buzz')
    else:
        print(str(i).zfill(3))
```

## Usage
```shell
$ go run  ./cmd/minivm/main.go run --link ./examples/ir/fizzbuzz/fizzbuzz.mir
```
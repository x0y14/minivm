package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/x0y14/minivm/asm"
	"github.com/x0y14/minivm/vm"
)

func readAsm(path string) (string, error) {
	if !strings.HasSuffix(path, ".mini") {
		return "", fmt.Errorf("error: unsupported file: %s", path)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func main() {
	var stackSize uint
	var heapSize uint

	cmd := &cli.Command{
		Name:  "minivm",
		Usage: "small stack machine",
		Commands: []*cli.Command{
			{
				Name:        "run",
				Usage:       "Execute program",
				Description: "execute .mini file as program",
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:        "stack",
						Value:       100,
						Usage:       "initial stack size",
						Destination: &stackSize,
					},
					&cli.UintFlag{
						Name:        "heap",
						Value:       100,
						Usage:       "initial heap size",
						Destination: &heapSize,
					},
				},
				Action: func(ctx context.Context, command *cli.Command) error {
					// load args
					var filePaths []string
					for i := 0; i < command.Args().Len(); i++ {
						filePaths = append(filePaths, command.Args().Get(i))
					}
					// check count of files
					if len(filePaths) <= 0 {
						// 最低一個は指定してね
						return fmt.Errorf("error: at least one file must be specified")
					}
					if len(filePaths) != 1 {
						// ファイル多すぎるべ
						return fmt.Errorf("error: too many files specified")
					}

					// read file
					assembly, err := readAsm(filePaths[0])
					if err != nil {
						return err
					}
					// tokenize
					tokens, err := asm.Tokenize([]rune(assembly))
					if err != nil {
						return err
					}
					// parse
					nodes, err := asm.Parse(tokens)
					if err != nil {
						return err
					}
					// gen code
					codes, err := asm.Gen(nodes)
					if err != nil {
						return err
					}

					// exec
					rt := vm.NewRuntime(codes, &vm.Config{
						StackSize: int(stackSize),
						HeapSize:  int(heapSize),
					})
					return rt.Run()
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

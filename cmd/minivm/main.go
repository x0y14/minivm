package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/x0y14/minivm/bytecode"
	"github.com/x0y14/minivm/ir"
	"github.com/x0y14/minivm/vm"
)

func readByt(path string) (string, error) {
	if !strings.HasSuffix(path, ".mbyt") {
		return "", fmt.Errorf("error: unsupported file: %s", path)
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func readIr(path string) (string, error) {
	if !strings.HasSuffix(path, ".mir") {
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
	var link bool

	cmd := &cli.Command{
		Name:  "minivm",
		Usage: "small stack machine",
		Commands: []*cli.Command{
			{
				Name:  "link",
				Usage: "Link *.mir files",
				Action: func(ctx context.Context, command *cli.Command) error {
					var filePaths []string
					for i := 0; i < command.Args().Len(); i++ {
						filePaths = append(filePaths, command.Args().Get(i))
					}
					// *.mir
					var irs []*ir.IR
					for _, filePath := range filePaths {
						v, err := readIr(filePath)
						if err != nil {
							return err
						}
						tokens, err := ir.Tokenize([]rune(v), true)
						if err != nil {
							return err
						}
						ir_, err := ir.Parse(tokens)
						if err != nil {
							return err
						}
						irs = append(irs, ir_)
					}
					nds, err := ir.Link(irs)
					if err != nil {
						return err
					}
					fmt.Printf(ir.Print(nds))
					return nil
				},
			},
			{
				Name:        "run",
				Usage:       "Execute program",
				Description: "execute .mbyt file as program",
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
					&cli.BoolFlag{
						Name:        "link",
						Aliases:     []string{"l"},
						Destination: &link,
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
					if !link && len(filePaths) != 1 {
						// リンクじゃないのにファイル多すぎるべ
						return fmt.Errorf("error: too many files specified")
					}

					var assembly string
					if link {
						// *.mir
						var irs []*ir.IR
						for _, filePath := range filePaths {
							v, err := readIr(filePath)
							if err != nil {
								return err
							}
							tokens, err := ir.Tokenize([]rune(v), true)
							if err != nil {
								return err
							}
							ir_, err := ir.Parse(tokens)
							if err != nil {
								return err
							}
							irs = append(irs, ir_)
						}
						nds, err := ir.Link(irs)
						if err != nil {
							return err
						}
						assembly = ir.Print(nds)
					} else {
						// *.mbyt
						// read file
						asm, err := readByt(filePaths[0])
						if err != nil {
							return err
						}
						assembly = asm
					}
					// tokenize
					tokens, err := bytecode.Tokenize([]rune(assembly))
					if err != nil {
						return err
					}
					// parse
					nodes, err := bytecode.Parse(tokens)
					if err != nil {
						return err
					}
					// gen code
					codes, err := bytecode.Gen(nodes)
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

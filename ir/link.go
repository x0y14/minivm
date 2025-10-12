package ir

import (
	"fmt"
	"strconv"
)

func adjustNode(n Node) Node {
	switch v := n.(type) {
	case Offset:
		if v.Target == PC {
			v.Diff += 2
		}
		return v
	case Instruction:
		newArgs := make([]Node, 0, len(v.Args))
		for _, a := range v.Args {
			newArgs = append(newArgs, adjustNode(a))
		}
		return Instruction{Op: v.Op, Args: newArgs}
	default:
		return n
	}
}

func merge(dst, src *IR) (*IR, error) {
	mergedImports := dst.Imports
	var mergedText = dst.Text
	geta := len(mergedText)
	for _, nd := range src.Text {
		switch nd := nd.(type) {
		case Offset:
			switch nd.Target { // PCの時だけdstのNODE分下駄を履かせる
			case PC:
				mergedText = append(mergedText, Offset{PC, nd.Diff + geta})
			default:
				mergedText = append(mergedText, nd)
			}
		case Label:
			if nd.Define {
				// mergedImportsから消す
				newImports := make([]string, 0, len(mergedImports))
				for _, imp := range mergedImports {
					if imp != nd.Name {
						newImports = append(newImports, imp)
					}
				}
				mergedImports = newImports
			}
			mergedText = append(mergedText, nd)
		default:
			mergedText = append(mergedText, nd)
		}
	}
	merged := &IR{
		Id:         dst.Id,
		EntryPoint: dst.EntryPoint,
		Imports:    mergedImports,
		Exports:    append(dst.Exports, src.Exports...),
		Constants:  append(dst.Constants, src.Constants...),
		Text:       mergedText,
	}
	merged.Imports = mergedImports
	return merged, nil
}

func solveData(ir *IR) ([]Node, error) {
	// constantの順番でheapの位置が解決できるはず
	var pre []Node
	var result []Node
	hp := 0
	for _, c := range ir.Constants {
		// heap位置をラベルと置換
		for _, nd := range ir.Text {
			switch nd := nd.(type) {
			case Label:
				if !nd.Define && nd.Name == c.Name {
					result = append(result, Number(hp))
				} else {
					result = append(result, nd)
				}
			default:
				result = append(result, nd)
			}
		}
		// preで使用するデータの作成
		pre = append(pre, ALLOC, Number(len(c.Values)), POP, R10)
		for i, v := range c.Values {
			switch v := v.(type) {
			case ConstChar:
				pre = append(pre, STORE, Number(hp+i), Character(v))
			case ConstInt:
				pre = append(pre, STORE, Number(hp+i), Number(v))
			}
		}

		hp += len(c.Values)
	}
	return pre, nil
}

// real entry point, error
func solve(ir *IR) (int, error) {
	preLocation := 0

	var preResult []Node
	labelLocations := map[string]int{}
	// ラベル定義の位置だけ全て取得する
	for pc, nd := range ir.Text {
		label, ok := nd.(Label)
		// ラベルでなければそのまま
		if !ok {
			preResult = append(preResult, nd)
			continue
		}
		// 定義かつexportされていなかったら
		if label.Define {
			labelLocations[label.Name] = pc
			// 無操作と入れ替える
			preResult = append(preResult, NOP)
			if label.Name == "_pre" {
				preLocation = pc
			}
			continue
		} else {
			// そのまま
			preResult = append(preResult, nd)
			continue
		}
	}

	var result []Node
	for pc, nd := range preResult {
		label, ok := nd.(Label)
		// ラベルでなければそのまま
		if !ok {
			result = append(result, nd)
			continue
		}
		// ラベル呼び出し かつ 場所が記録されている
		dst, ok := labelLocations[label.Name]
		if !label.Define && ok {
			// pc-1なのはjmpなどのOPが基準になるから
			result = append(result, Offset{PC, dst - (pc - 1)})
			continue
		} else {
			result = append(result, nd)
			continue
		}
	}

	ir.Text = result
	return preLocation, nil
}

func Link(irs []*IR) ([]Node, error) {
	globalTable := &SymbolTable{"global", make(map[string]Symbol)}
	// ラベル解決
	var entryPoint string
	entryCount := 0
	irMap := map[string]*IR{}
	for i, ir := range irs {
		// 適当にIRに名前をつける
		ir.Id = strconv.Itoa(i)
		if ir.EntryPoint != "" {
			entryCount++
			entryPoint = ir.EntryPoint
		}
		irMap[ir.Id] = ir
	}
	// エントリーポイントが複数存在しないかチェック
	if entryCount > 1 {
		return nil, fmt.Errorf("too many entryPoint: %d", entryCount)
	}

	resultIr := &IR{
		Id:         "",
		Imports:    []string{},
		Exports:    []string{},
		Constants:  []Constant{},
		EntryPoint: entryPoint,
		Text:       []Node{},
	}
	for _, ir := range irs {
		mergedIr, err := merge(resultIr, ir)
		if err != nil {
			return nil, err
		}
		resultIr = mergedIr
		if err := globalTable.collect(ir); err != nil {
			return nil, err
		}
		if len(globalTable.undefined()) > 0 {
			return nil, fmt.Errorf("undefined: %v", globalTable.undefined())
		}
	}

	//if err := globalTable.collect(resultIr); err != nil {
	//	return nil, err
	//}
	unsolved := globalTable.unsolved()
	if len(unsolved) > 0 {
		return nil, fmt.Errorf("unsolved label exists: %v", unsolved)
	}

	// sizeofを解決する
	_, nds, err := solveSizeof([]string{}, resultIr.Constants, resultIr.Text)
	if err != nil {
		return nil, err
	}
	// 定数解決
	preScript, err := solveData(resultIr)
	if err != nil {
		return nil, err
	}

	resultIr.Text = nds
	resultIr.Text = append(resultIr.Text, []Node{
		Label{Define: true, Name: "_pre"},
	}...)
	resultIr.Text = append(resultIr.Text, preScript...)
	resultIr.Text = append(resultIr.Text, JMP, Label{false, "_start"})

	vmEntryPoint, err := solve(resultIr)
	if err != nil {
		return nil, err
	}
	resultIr.Text = append([]Node{JMP, Number(vmEntryPoint)}, resultIr.Text...)
	// 全てのラベル位置を+2する
	adjusted := make([]Node, 0, len(resultIr.Text))
	for _, nd := range resultIr.Text {
		adjusted = append(adjusted, adjustNode(nd))
	}

	return adjusted, nil
}

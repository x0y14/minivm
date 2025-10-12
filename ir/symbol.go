package ir

import "fmt"

type SymbolKind int

const (
	_        SymbolKind = iota
	Data                // eg) msg
	Function            // eg) _start
	//Local               // eg) __loop_start
	Unknown
	Undefined
)

type Symbol struct {
	Kind     SymbolKind
	Name     string
	Source   string
	Exported bool
}

type SymbolTable struct {
	Id      string
	Symbols map[string]Symbol
}

func (s *SymbolTable) declare(kind SymbolKind, name string, source string, exported bool) error {
	for exists, sym := range s.Symbols {
		if exists == name { // あった
			// 一時データでは上書きしない
			if kind == Undefined || kind == Unknown {
				return nil
			}
			if sym.Kind == Undefined {
				delete(s.Symbols, name)
				// 登録
				s.Symbols[name] = Symbol{
					Kind:     kind,
					Name:     name,
					Source:   source,
					Exported: exported,
				}
				return nil
			} else if sym.Kind == Unknown {
				if !exported {
					return fmt.Errorf("unexported overwrite")
				}
				delete(s.Symbols, name)
				// 登録
				s.Symbols[name] = Symbol{
					Kind:     kind,
					Name:     name,
					Source:   source,
					Exported: exported,
				}
				return nil
			}
			return fmt.Errorf("label exists: %s", name)
		}
	}
	s.Symbols[name] = Symbol{
		Kind:     kind,
		Name:     name,
		Source:   source,
		Exported: exported,
	}
	return nil
}

func (s *SymbolTable) pull(entryPoint string, table *SymbolTable) error {
	for _, sym := range table.Symbols {
		if sym.Exported || sym.Name == entryPoint {
			if err := s.declare(sym.Kind, sym.Name, sym.Source, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SymbolTable) undefined() []Symbol {
	var undef []Symbol
	for _, sym := range s.Symbols {
		if sym.Kind == Undefined {
			undef = append(undef, sym)
		}
	}
	return undef
}

func (s *SymbolTable) unsolved() []Symbol {
	var unsolved_ []Symbol
	for _, sym := range s.Symbols {
		if sym.Kind == Unknown || sym.Kind == Undefined {
			unsolved_ = append(unsolved_, sym)
		}
	}
	return unsolved_
}

func (s *SymbolTable) collect(ir *IR) error {
	for _, label := range ir.Imports {
		if err := s.declare(Unknown, label, ir.Id, false); err != nil {
			return err
		}
	}
	for _, c := range ir.Constants {
		if err := s.declare(Data, c.Name, ir.Id, in(c.Name, ir.Exports)); err != nil {
			return err
		}
	}
	for _, nd := range ir.Text {
		switch nd := nd.(type) {
		case Label:
			if nd.Define {
				if err := s.declare(Function, nd.Name, ir.Id, in(nd.Name, ir.Exports)); err != nil {
					return err
				}
			} else {
				// 未定義（exports に含まれていれば exported=true として登録）
				if err := s.declare(Undefined, nd.Name, ir.Id, in(nd.Name, ir.Exports)); err != nil {
					return err
				}
			}
		default:
			// label以外は無視
		}
	}
	return nil
}

package asm

import (
	"fmt"
	"strconv"
)

var text []rune
var loc location

type location struct {
	at       int
	line     int
	atInLine int
}

type TokenKind int

const (
	_ TokenKind = iota
	Eof
	Comment

	Identifier
	Integer
	String

	Lrb // (
	Rrb // )
	Lcb // [
	Rcb // ]
	At  // @

	Dot   // .
	Comma // ,
	Colon // :

	Add // +
	Sub // -
	Mul // *
)

func (tk TokenKind) String() string {
	kinds := []string{
		Eof:        "Eof",
		Comment:    "Comment",
		Identifier: "Identifier",
		Integer:    "Integer",
		String:     "String",
		Lrb:        "(",
		Rrb:        ")",
		Lcb:        "[",
		Rcb:        "]",
		At:         "@",
		Dot:        ".",
		Colon:      ":",
		Add:        "+",
		Mul:        "*",
	}
	return kinds[tk]
}

type Token struct {
	Kind     TokenKind
	Position Position
	Raw      []rune
	Next     *Token
}

func (t *Token) GetValueAsInteger() (int, error) {
	if t.Kind != Integer {
		return 0, fmt.Errorf("type mismatch: actual=%s", t.Kind.String())
	}
	i64, err := strconv.ParseInt(string(t.Raw), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}

func (t *Token) GetValueAsString() (string, error) {
	if t.Kind != String {
		return "", fmt.Errorf("type mismatch: actual=%s", t.Kind.String())
	}
	return string(t.Raw), nil
}

func comment() (*Token, error) {
	tok := Token{Kind: Comment, Position: Position{loc.atInLine, loc.line}}
	v := ""
	for loc.at < len(text) && text[loc.at] != '\n' {
		v += string(text[loc.at])
		loc.at++
		loc.atInLine++
	}
	tok.Raw = []rune(v)
	return &tok, nil
}

func isIdentifier(head bool, r rune) bool {
	lower := 'a' <= r && r <= 'z'
	upper := 'A' <= r && r <= 'Z'
	sym := '_' == r
	num := isNumeric(r)
	if head {
		return lower || upper || sym
	}
	return lower || upper || sym || num
}

func identifier() (*Token, error) {
	tok := Token{Kind: Identifier, Position: Position{StartedAt: loc.atInLine, Line: loc.line}}
	v := ""
	for loc.at < len(text) && isIdentifier(false, text[loc.at]) {
		v += string(text[loc.at])
		loc.at++
		loc.atInLine++
	}
	tok.Raw = []rune(v)
	return &tok, nil
}

func isNumeric(r rune) bool {
	return '0' <= r && r <= '9'
}

func integer() (*Token, error) {
	tok := Token{Kind: Integer, Position: Position{loc.atInLine, loc.line}}
	v := ""
	for loc.at < len(text) && isNumeric(text[loc.at]) {
		v += string(text[loc.at])
		loc.at++
		loc.atInLine++
	}
	tok.Raw = []rune(v)
	return &tok, nil
}

func isSymbol(r rune) bool {
	return r == '(' || r == ')' || r == '[' || r == ']' ||
		r == '@' ||
		r == '.' || r == ',' || r == ':' ||
		r == '+' || r == '-' || r == '*'
}

func symbol() (*Token, error) {
	sym := map[rune]Token{
		'(': {Kind: Lrb},
		')': {Kind: Rrb},
		'[': {Kind: Lcb},
		']': {Kind: Rcb},
		'@': {Kind: At},
		'.': {Kind: Dot},
		',': {Kind: Comma},
		':': {Kind: Colon},
		'+': {Kind: Add},
		'-': {Kind: Sub},
		'*': {Kind: Mul},
	}
	tok, ok := sym[text[loc.at]]
	if !ok {
		return nil, fmt.Errorf("unexpected rune: %s", string(text[loc.at]))
	}
	tok.Position = Position{StartedAt: loc.atInLine, Line: loc.line}
	loc.at++
	loc.atInLine++
	return &tok, nil
}

func Tokenize(input []rune) (*Token, error) {
	text = input
	loc = location{0, 0, 0}
	head := &Token{}
	curt := head

	for loc.at < len(text) {
		switch r := text[loc.at]; {
		case r == ' ' || r == '\t':
			loc.at++
			loc.atInLine++
		case r == '\n':
			loc.at++
			loc.line++
			loc.atInLine = 0
		case r == ';':
			tok, err := comment()
			if err != nil {
				return nil, err
			}
			curt.Next = tok
			curt = curt.Next
		case isIdentifier(true, r):
			tok, err := identifier()
			if err != nil {
				return nil, err
			}
			curt.Next = tok
			curt = curt.Next
		case isNumeric(r):
			tok, err := integer()
			if err != nil {
				return nil, err
			}
			curt.Next = tok
			curt = curt.Next
		case isSymbol(r):
			tok, err := symbol()
			if err != nil {
				return nil, err
			}
			curt.Next = tok
			curt = curt.Next
		default:
			return nil, fmt.Errorf("unexpected rune: %s", string(text[loc.at]))
		}
	}
	curt.Next = &Token{
		Kind:     Eof,
		Position: Position{StartedAt: loc.atInLine, Line: loc.line},
	}
	return head.Next, nil
}

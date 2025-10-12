package ir

import "strings"

func Print(nodes []Node) string {
	nodes = expand(nodes)
	out := ""
	pos := 0
	for pos < len(nodes) {
		switch node := nodes[pos].(type) {
		case Operation:
			line := []string{node.String()}
			pos++
			for i := 0; i < node.NumOperands(); i++ {
				line = append(line, nodes[pos].String())
				pos++
			}
			out += strings.Join(line, " ") + "\n"
		default:
			out += node.String() + "\n"
			pos++
		}
	}
	return out
}

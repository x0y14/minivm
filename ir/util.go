package ir

func in(target string, arr []string) bool {
	for _, elm := range arr {
		if elm == target {
			return true
		}
	}
	return false
}

func expand(nodes []Node) []Node {
	var nds []Node
	for _, n := range nodes {
		switch n := n.(type) {
		case Instruction:
			nds = append(nds, n.Nodes()...)
		default:
			nds = append(nds, n)
		}
	}
	return nds
}

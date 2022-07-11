package ast

import (
	"os"
	"fmt"
)

type Node interface {
    Readable(indent int) string
    At() string
    End() string
}

var ast []Decl // only declaring/defining variables/functions allowed in global scope

func ShowAst() {
    res := ""
    for _, node := range ast {
        res += node.Readable(0)
    }

    fmt.Print(res);
}

func AddNode(decl Decl) {
    ast = append(ast, decl)
}

func Compile(asm *os.File) {
    for _, node := range ast {
        node.Compile(asm)
    }
}

package ast

import (
	"fmt"
)

type Node interface {
    Readable(indent int) string
    At() string
    End() string
}

type Ast struct {
    Decls []Decl // only declaring/defining variables/functions allowed in global scope
}

func (ast *Ast) ShowAst() {
    res := ""
    for _, node := range ast.Decls {
        res += node.Readable(0)
    }

    fmt.Print(res);
}

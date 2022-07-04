package ast

import (
	"os"
	"fmt"
)

var ast OpProgramm

func ShowAst() {
    fmt.Print(ast.Readable(0));
}

func AddOp(opDecl OpDecl) {
    ast.Ops = append(ast.Ops, opDecl)
}

func Compile(asm *os.File) {
    ast.Compile(asm);
}


// TODO At() for Op
type Op interface {
    Readable(indent int) string
}

type OpProgramm struct {
    Ops []OpDecl // only declaring/defining variables/functions allowed in global scope
}

func (o *OpProgramm) Readable(indent int) string {
    res := ""
    for _, op := range o.Ops {
        res += op.Readable(indent)
    }

    return res
}

func (o *OpProgramm) Compile(asm *os.File) {
    for _, op := range o.Ops {
        op.Compile(asm)
    }
}

package ast

import (
    "fmt"
    "os"
)

var ast OpProgramm

func ShowAst() {
    fmt.Println(ast.Readable(0));
}

func AddOp(opDecl OpDecl) {
    ast.Ops = append(ast.Ops, opDecl)
}

func Compile(asm *os.File) {
    ast.Compile(asm);
}


type OpType uint
const (
    OP_DEC_VAR  OpType = iota
    OP_DEF_VAR
    OP_DEF_FN
    OP_CALL_FN
    OP_BLOCK
    OP_COUNT      uint = iota
)

func (o OpType) Readable() string {
    // compile time reminder to add cases when Operants are added
    const _ uint = 5 - OP_COUNT

    switch o {
    case OP_DEC_VAR:
        return "OP_DEC_VAR"
    case OP_DEF_VAR:
        return "OP_DEF_VAR"
    case OP_DEF_FN:
        return "OP_DEF_FN"
    case OP_BLOCK:
        return "OP_BLOCK"
    case OP_CALL_FN:
        return "OP_CALL_FN"
    default:
        return ""
    }
}


type Op interface {
    Readable(indent int) string
    Compile(asm *os.File)
}

type OpProgramm struct {
    Ops []interface{ OpDecl } // only declaring/defining variables/functions allowed in global scope
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

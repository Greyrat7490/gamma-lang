package ast

import (
    "fmt"
    "os"
)

var Ast OpProgramm

func ShowAst() {
    fmt.Println(Ast.Readable(0));
}


type OpType uint
const (
    OP_DEC_VAR  OpType = iota
    OP_DEF_VAR
    OP_DEF_FN
    OP_END_FN
    OP_CALL_FN
    OP_DEC_ARGS
    OP_DEF_ARGS
    OP_COUNT      uint = iota
)

func (o OpType) Readable() string {
    // compile time reminder to add cases when Operants are added
    const _ uint = 7 - OP_COUNT

    switch o {
    case OP_DEC_VAR:
        return "OP_DEC_VAR"
    case OP_DEF_VAR:
        return "OP_DEF_VAR"
    case OP_DEF_FN:
        return "OP_DEF_FN"
    case OP_END_FN:
        return "OP_END_FN"
    case OP_CALL_FN:
        return "OP_CALL_FN"
    case OP_DEC_ARGS:
        return "OP_DEC_ARGS"
    case OP_DEF_ARGS:
        return "OP_DEF_ARGS"
    default:
        return ""
    }
}


type Op interface {
    Readable(indent int) string
    Compile(asm *os.File)
}

type OpProgramm struct {
    Ops []interface{ Op } // TODO: later only OpDecVar, OpDefVar, OpDefFn
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

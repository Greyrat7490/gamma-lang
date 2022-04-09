package ast

import (
    "os"
    "fmt"
)

var Ast []interface{ Op }

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
    Readable() string
    Compile(asm *os.File)
}

func ShowAst() {
    fmt.Println("AST START")

    for i, op := range Ast {
        fmt.Printf("%d: %s\n", i, op.Readable())
    }

    fmt.Println("AST END")
}

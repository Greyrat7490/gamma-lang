package prs

import (
    "fmt"
    "os"
    "gorec/token"
)

var Ops []interface{ Op }
var isMainDefined bool = false

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

func ShowOps() {
    for i, o := range Ops {
        fmt.Printf("%d: %s\n", i, o.Readable())
    }
}

func Parse() {
    tokens := token.GetTokens()

    for i := 0; i < len(tokens); i++ {
        switch tokens[i].Type {
        case token.Dec_var:
            i = prsDecVar(i)
        case token.Def_var:
            i = prsDefVar(i)
        case token.Def_fn:
            i = prsDefFn(i)
        case token.Name:
            if tokens[i+1].Type == token.ParenL {
                fmt.Fprintln(os.Stderr, "[ERROR] function calls are not allowed in global scope")
                fmt.Fprintln(os.Stderr, "\t" + tokens[i].At())
                os.Exit(1)
            }
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", tokens[i].Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens[i].At())
            os.Exit(1)
        }
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }
}

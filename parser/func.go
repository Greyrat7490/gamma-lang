package prs

import (
    "os"
    "fmt"
    "strings"
    "gorec/func"
    "gorec/token"
    "gorec/types"
    "gorec/ast"
)

type OpDefFn struct {
    FnName token.Token
    DecArgsOp OpDecArgs
    Ops []interface{ ast.Op }   // TODO: later all Ops except OpDefFn
}
func (o OpDefFn) Readable(indent int) string {
    res := strings.Repeat("   ", indent) +
        fmt.Sprintf("OP_DEF_FN:  %s (%s)\n",
        o.FnName.Str, o.FnName.Type.Readable()) +
        o.DecArgsOp.Readable(indent+1)

    for _, op := range o.Ops {
        res += op.Readable(indent+1)
    }

    return res
}
func (o OpDefFn) Compile(asm *os.File) {
    fn.Define(asm, o.FnName)
    o.DecArgsOp.Compile(asm)

    for _, op := range o.Ops {
        op.Compile(asm)
    }

    fn.End(asm);
}

type OpDecArgs struct {
    Args []fn.Arg
}
func (o OpDecArgs) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("OP_DEC_ARGS: %v\n", o.Args)
}
func (o OpDecArgs) Compile(asm *os.File) {
    fn.DeclareArgs(o.Args)
}


type OpFnCall struct {
    FnName token.Token
    DefArgsOp OpDefArgs
}
func (o OpFnCall) Readable(indent int) string {
    return strings.Repeat("   ", indent) +
        fmt.Sprintf("OP_CALL_FN: %s\n", o.FnName.Str) +
        o.DefArgsOp.Readable(indent+1)
}
func (o OpFnCall) Compile(asm *os.File) {
    o.DefArgsOp.Compile(asm)
    fn.CallFunc(asm, o.FnName)
}


type OpDefArgs struct {
    FnName token.Token
    Values []string
}
func (o OpDefArgs) Readable(indent int) string {
    return strings.Repeat("   ", indent) + fmt.Sprintf("OP_DEF_ARGS: %s %v\n", o.FnName.Str, o.Values)
}
func (o OpDefArgs) Compile(asm *os.File) {
    fn.DefineArgs(asm, o.FnName, o.Values)
}


func prsDefFn(idx int) (OpDefFn, int) {
    tokens := token.GetTokens()

    if tokens[idx+1].Str == "main" {
        isMainDefined = true
    }

    var op OpDefFn = OpDefFn{ FnName: tokens[idx+1] }

    op.DecArgsOp, idx = prsDecArgs(idx)

    for ; idx < len(tokens); idx++ {
        switch tokens[idx].Type {
        case token.Dec_var:
            var decOp OpDecVar
            decOp, idx = prsDecVar(idx)
            op.Ops = append(op.Ops, decOp)
        case token.Def_var:
            var defOp OpDefVar
            defOp, idx = prsDefVar(idx)
            op.Ops = append(op.Ops, defOp)
        case token.Def_fn:
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
        case token.BraceR:
            return op, idx
        case token.Name:
            if tokens[idx+1].Type == token.ParenL {
                var callOp OpFnCall
                callOp, idx = prsCallFn(idx)
                op.Ops = append(op.Ops, callOp)
            }
            // TODO: assign
        default:
            // TODO
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\"\n", tokens[idx].Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
        }
    }

    fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" was not closed (missing \"}\")\n", tokens[idx+1].Str)
    fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
    os.Exit(1)

    return OpDefFn{}, -1
}

func prsDecArgs(idx int) (OpDecArgs, int) {
    tokens := token.GetTokens()

    if len(tokens) < idx + 2 {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    if tokens[idx+2].Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %s(\"%s\")\n", tokens[idx+2].Type.Readable(), tokens[idx+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+2].At())
        os.Exit(1)
    }

    var op OpDecArgs

    var a fn.Arg
    b := false
    for _, w := range tokens[idx+3:] {
        if w.Type == token.ParenR {
            b = true
            break
        }

        if w.Type == token.BraceL || w.Type == token.BraceR {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }

        if a.Name == "" {
            if w.Type != token.Name {
                fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %s(\"%s\")\n", w.Type.Readable(), w.Str)
                fmt.Fprintln(os.Stderr, "\t" + w.At())
                os.Exit(1)
            }

            a.Name = w.Str
        } else {
            if w.Type != token.Typename {
                fmt.Fprintf(os.Stderr, "[ERROR] expected a Typename but got %s(\"%s\")\n", w.Type.Readable(), w.Str)
                fmt.Fprintln(os.Stderr, "\t" + w.At())
                os.Exit(1)
            }

            a.Type = types.ToType(w.Str)
            op.Args = append(op.Args, a)

            a.Name = ""
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\" for function \"%s\"\n", tokens[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }

    return op, idx + len(op.Args) * 2 + 5
}

func prsCallFn(idx int) (OpFnCall, int) {
    tokens := token.GetTokens()

    var op OpFnCall = OpFnCall{ FnName: tokens[idx] }

    if len(tokens) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }
    if tokens[idx+1].Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %s(\"%s\")\n", tokens[idx+1].Type.Readable(), tokens[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }

    op.DefArgsOp, idx = prsDefArgs(idx)

    return op, idx
}

func prsDefArgs(idx int) (OpDefArgs, int) {
    tokens := token.GetTokens()

    var op OpDefArgs = OpDefArgs{ FnName: tokens[idx] }

    b := false
    for _, w := range tokens[idx+2:] {
        if w.Type == token.ParenR {
            b = true
            break
        }

        if w.Type == token.BraceL || w.Type == token.BraceR {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }

        if !(w.Type == token.Number || w.Type == token.Str || w.Type == token.Name) {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Name or a literal but got %s(\"%s\")\n", w.Type.Readable(), w.Str)
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }

        op.Values = append(op.Values, w.Str)
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    return op, idx + len(op.Values) + 2
}

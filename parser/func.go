package prs

import (
    "os"
    "fmt"
    "gorec/func"
    "gorec/token"
    "gorec/types"
)

type OpDefFn struct {
    FnName token.Token
}
func (o OpDefFn) Readable() string {
    return fmt.Sprintf("%s: %s", OP_DEF_FN.Readable(), o.FnName)
}
func (o OpDefFn) Compile(asm *os.File) {
    fn.Define(asm, o.FnName)
}

type OpEndFn struct {}
func (o OpEndFn) Readable() string {
    return fmt.Sprintf("%s", OP_END_FN.Readable())
}
func (o OpEndFn) Compile(asm *os.File) {
    fn.End(asm)
}


type OpDecArgs struct {
    Args []fn.Arg
}
func (o OpDecArgs) Readable() string {
    return fmt.Sprintf("%s: %v", OP_DEF_ARGS.Readable(), o.Args)
}
func (o OpDecArgs) Compile(asm *os.File) {
    fn.DeclareArgs(o.Args)
}


type OpFnCall struct {
    FnName token.Token
}
func (o OpFnCall) Readable() string {
    return fmt.Sprintf("%s: %s", OP_CALL_FN.Readable(), o.FnName)
}
func (o OpFnCall) Compile(asm *os.File) {
    fn.CallFunc(asm, o.FnName)
}


type OpDefArgs struct {
    FnName token.Token
    Values []string
}
func (o OpDefArgs) Readable() string {
    return fmt.Sprintf("%s: %s %v", OP_DEF_ARGS.Readable(), o.FnName, o.Values)
}
func (o OpDefArgs) Compile(asm *os.File) {
    fn.DefineArgs(asm, o.FnName, o.Values)
}


func prsDefFn(idx int) int {
    tokens := token.GetTokens()

    if tokens[idx+1].Str == "main" {
        isMainDefined = true
    }

    var op OpDefFn = OpDefFn{ FnName: tokens[idx+1] }
    Ops = append(Ops, op)

    idx = prsDecArgs(idx)

    for ; idx < len(tokens); idx++ {
        switch tokens[idx].Type {
        case token.Dec_var:
            idx = prsDecVar(idx)
        case token.Def_var:
            idx = prsDefVar(idx)
        case token.Def_fn:
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
        case token.BraceR:
            prsEnd()
            return idx
        case token.Name:
            if tokens[idx+1].Type == token.ParenL {
                idx = prsCallFn(idx)
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

    return -1
}

func prsEnd() {
    Ops = append(Ops, OpEndFn{})
}

func prsDecArgs(idx int) int {
    tokens := token.GetTokens()

    if len(tokens) < idx + 2 || tokens[idx+2].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+2].At())
        os.Exit(1)
    }

    var op OpDecArgs

    var a fn.Arg
    b := false
    for _, w := range tokens[idx+3:] {
        if w.Str == ")" {
            b = true
            break
        }

        if w.Str == "{" || w.Str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }

        if a.Name == "" {
            a.Name = w.Str
        } else {
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

    if len(op.Args) > 0 {
        Ops = append(Ops, op)
    }

    return idx + len(op.Args) * 2 + 5
}

func prsCallFn(idx int) int {
    tokens := token.GetTokens()

    var op OpFnCall = OpFnCall{ FnName: tokens[idx] }

    if len(tokens) < idx + 1 || tokens[idx+1].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }

    idx = prsDefArgs(idx)

    Ops = append(Ops, op)

    return idx
}

func prsDefArgs(idx int) int {
    tokens := token.GetTokens()

    var op OpDefArgs = OpDefArgs{ FnName: tokens[idx] }

    b := false
    for _, w := range tokens[idx+2:] {
        if w.Str == ")" {
            b = true
            break
        }

        if w.Str == "{" || w.Str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }

        op.Values = append(op.Values, w.Str)
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    Ops = append(Ops, op)

    return idx + len(op.Values) + 2
}

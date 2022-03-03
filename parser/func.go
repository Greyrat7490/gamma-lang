package prs

import (
    "fmt"
    "os"
    "gorec/types"
    "gorec/func"
)

type OpDefFn struct {
    Funcname string
}
func (o OpDefFn) Readable() string {
    return fmt.Sprintf("%s: %s", OP_DEF_FN.Readable(), o.Funcname)
}
func (o OpDefFn) Compile(asm *os.File) {
    fn.Define(asm, o.Funcname)
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
    FnName string
}
func (o OpFnCall) Readable() string {
    return fmt.Sprintf("%s: %s", OP_CALL_FN.Readable(), o.FnName)
}
func (o OpFnCall) Compile(asm *os.File) {
    fn.CallFunc(asm, o.FnName)
}


type OpDefArgs struct {
    FnName string
    Values []string
}
func (o OpDefArgs) Readable() string {
    return fmt.Sprintf("%s: %s %v", OP_DEF_ARGS.Readable(), o.FnName, o.Values)
}
func (o OpDefArgs) Compile(asm *os.File) {
    fn.DefineArgs(asm, o.FnName, o.Values)
}


func prsDefFn(words []Token, idx int) int {
    if words[idx+1].Str == "main" {
        isMainDefined = true
    }

    var op OpDefFn = OpDefFn{ Funcname: words[idx+1].Str }
    Ops = append(Ops, op)

    idx = prsDecArgs(words, idx)

    for ; idx < len(words); idx++ {
        switch words[idx].Type {
        case dec_var:
            idx = prsDecVar(words, idx)
        case def_var:
            idx = prsDefVar(words, idx)
        case def_fn:
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
            os.Exit(1)
        case braceR:
            prsEnd(words, idx)
            return idx
        case name:
            if tokens[idx+1].Type == parenL {
                idx = prsCallFn(words, idx)
            }
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\"\n", words[idx].Str)
            fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
            os.Exit(1)
        }
    }

    fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" was not closed (missing \"}\")\n", words[idx+1].Str)
    fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
    os.Exit(1)

    return -1
}

func prsEnd(words []Token, idx int) {
    Ops = append(Ops, OpEndFn{})
}

func prsDecArgs(words []Token, idx int) int {
    if len(words) < idx + 2 || words[idx+2].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].At())
        os.Exit(1)
    }

    var op OpDecArgs

    var a fn.Arg
    b := false
    for _, w := range words[idx+3:] {
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
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\" for function \"%s\"\n", words[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    if len(op.Args) > 0 {
        Ops = append(Ops, op)
    }

    return idx + len(op.Args) * 2 + 5
}

func prsCallFn(words []Token, idx int) int {
    var op OpFnCall = OpFnCall{ FnName: words[idx].Str }

    if len(words) < idx + 1 || words[idx+1].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    idx = prsDefArgs(words, idx)

    Ops = append(Ops, op)

    return idx
}

func prsDefArgs(words []Token, idx int) int {
    var op OpDefArgs = OpDefArgs{ FnName: words[idx].Str }

    b := false
    for _, w := range words[idx+2:] {
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

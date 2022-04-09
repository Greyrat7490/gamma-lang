package prs

import (
    "os"
    "fmt"
    "gorec/func"
    "gorec/token"
    "gorec/types"
    "gorec/ast"
)

type OpDefFn struct {
    FnName token.Token
}
func (o OpDefFn) Readable() string {
    return fmt.Sprintf("%s: %s", ast.OP_DEF_FN.Readable(), o.FnName.Str)
}
func (o OpDefFn) Compile(asm *os.File) {
    fn.Define(asm, o.FnName)
}

type OpEndFn struct {}
func (o OpEndFn) Readable() string {
    return fmt.Sprintf("%s", ast.OP_END_FN.Readable())
}
func (o OpEndFn) Compile(asm *os.File) {
    fn.End(asm)
}


type OpDecArgs struct {
    Args []fn.Arg
}
func (o OpDecArgs) Readable() string {
    return fmt.Sprintf("%s: %v", ast.OP_DEF_ARGS.Readable(), o.Args)
}
func (o OpDecArgs) Compile(asm *os.File) {
    fn.DeclareArgs(o.Args)
}


type OpFnCall struct {
    FnName token.Token
}
func (o OpFnCall) Readable() string {
    return fmt.Sprintf("%s: %s", ast.OP_CALL_FN.Readable(), o.FnName.Str)
}
func (o OpFnCall) Compile(asm *os.File) {
    fn.CallFunc(asm, o.FnName)
}


type OpDefArgs struct {
    FnName token.Token
    Values []string
}
func (o OpDefArgs) Readable() string {
    return fmt.Sprintf("%s: %s %v", ast.OP_DEF_ARGS.Readable(), o.FnName.Str, o.Values)
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
    ast.Ast = append(ast.Ast, op)

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
    ast.Ast = append(ast.Ast, OpEndFn{})
}

func prsDecArgs(idx int) int {
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

    if len(op.Args) > 0 {
        ast.Ast = append(ast.Ast, op)
    }

    return idx + len(op.Args) * 2 + 5
}

func prsCallFn(idx int) int {
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

    idx = prsDefArgs(idx)

    ast.Ast = append(ast.Ast, op)

    return idx
}

func prsDefArgs(idx int) int {
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

    ast.Ast = append(ast.Ast, op)

    return idx + len(op.Values) + 2
}

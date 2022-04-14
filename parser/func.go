package prs

import (
    "os"
    "fmt"
    "gorec/func"
    "gorec/token"
    "gorec/types"
    "gorec/ast"
)

func prsDefFn(idx int) (ast.OpDefFn, int) {
    tokens := token.GetTokens()

    if tokens[idx+1].Str == "main" {
        isMainDefined = true
    }

    var op ast.OpDefFn = ast.OpDefFn{ FnName: tokens[idx+1] }

    op.Args, idx = prsDecArgs(idx)

    for ; idx < len(tokens); idx++ {
        switch tokens[idx].Type {
        case token.Dec_var:
            var decOp ast.OpDecVar
            decOp, idx = prsDecVar(idx)
            op.Block.Stmts = append(op.Block.Stmts, ast.OpDeclStmt{ Decl: decOp })
        case token.Def_var:
            var defOp ast.OpDefVar
            defOp, idx = prsDefVar(idx)
            op.Block.Stmts = append(op.Block.Stmts, ast.OpDeclStmt{ Decl: defOp })
        case token.Def_fn:
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
        case token.BraceR:
            return op, idx
        case token.Name:
            if tokens[idx+1].Type == token.ParenL {
                var callOp ast.OpFnCall
                callOp, idx = prsCallFn(idx)
                op.Block.Stmts = append(op.Block.Stmts, ast.OpExprStmt{ Expr: callOp })
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

    return ast.OpDefFn{}, -1
}

func prsDecArgs(idx int) ([]fn.Arg, int) {
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

    var args []fn.Arg

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
            args = append(args, a)

            a.Name = ""
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\" for function \"%s\"\n", tokens[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }

    return args, idx + len(args) * 2 + 5
}

func prsCallFn(idx int) (ast.OpFnCall, int) {
    tokens := token.GetTokens()

    var op ast.OpFnCall = ast.OpFnCall{ FnName: tokens[idx] }

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

    op.Values, idx = prsDefArgs(idx)

    return op, idx
}

func prsDefArgs(idx int) ([]string, int) {
    tokens := token.GetTokens()

    var values []string

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

        values = append(values, w.Str)
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    return values, idx + len(values) + 2
}

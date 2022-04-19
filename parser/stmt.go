package prs

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/ast"
)

func prsStmt(idx int) (ast.OpStmt, int) {
    tokens := token.GetTokens()

    switch tokens[idx].Type {
    case token.Dec_var:
        var decOp ast.OpDecVar
        decOp, idx = prsDecVar(idx)
        return &ast.OpDeclStmt{ Decl: &decOp }, idx

    case token.Def_var:
        var defOp ast.OpDefVar
        defOp, idx = prsDefVar(idx)
        return &ast.OpDeclStmt{ Decl: &defOp }, idx

    case token.Name:
        if tokens[idx+1].Type == token.ParenL {
            var callOp ast.OpFnCall
            callOp, idx = prsCallFn(idx)
            return &ast.OpExprStmt{ Expr: &callOp }, idx
        } else if tokens[idx+1].Type == token.Assign {
            var o ast.OpAssignVar
            o, idx = prsAssignVar(idx+1)
            return &o, idx
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not used\n", tokens[idx].Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
            return &ast.BadStmt{}, -1
        }

    case token.Assign:
        fmt.Fprintf(os.Stderr, "[ERROR] no destination for assignment\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.BadStmt{}, -1

    case token.Def_fn:
        fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.BadStmt{}, -1

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected token \"%s\" (of type \"%s\")\n", tokens[idx].Str, tokens[idx].Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.BadStmt{}, -1
    }
}

func prsAssignVar(idx int) (ast.OpAssignVar, int) {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    v := tokens[idx-1]
    value, idx := prsExpr(idx+1)

    op := ast.OpAssignVar{ Varname: v, Value: value }

    return op, idx
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

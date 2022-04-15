package prs

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/ast"
)


func prsDecVar(idx int) (ast.OpDecVar, int) {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] neither name nor type provided for the variable declaration")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }
    if len(tokens) < idx + 2 {
        if tokens[idx+1].Type == token.Name {
            fmt.Fprintln(os.Stderr, "[ERROR] no type provided for the variable")
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] no name provided for the variable")
        }
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }

    if (tokens[idx+1].Type != token.Name) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %s(\"%s\")\n", tokens[idx+1].Type.Readable(), tokens[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }
    if (tokens[idx+2].Type != token.Typename) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Typename but got %s(\"%s\")\n", tokens[idx+2].Type.Readable(), tokens[idx+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+2].At())
        os.Exit(1)
    }


    t := types.ToType(tokens[idx+2].Str)
    if t == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", tokens[idx+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+2].At())
        os.Exit(1)
    }

    op := ast.OpDecVar{ Varname: tokens[idx+1], Vartype: t }

    return op, idx + 2
}

func prsDefVar(idx int) (ast.OpDefVar, int) {
    tokens := token.GetTokens()

    if len(tokens) < idx + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    if (tokens[idx-2].Type != token.Name) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %s(\"%s\")\n", tokens[idx-2].Type.Readable(), tokens[idx-2].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx-2].At())
        os.Exit(1)
    }

    name := tokens[idx-2]
    value, idx := prsExpr(idx+1)

    if (!(tokens[idx].Type == token.Name || tokens[idx].Type == token.Number || tokens[idx].Type == token.Str)) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name or a literal but got %s(\"%s\")\n", tokens[idx].Type.Readable(), tokens[idx].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    op := ast.OpDefVar{ Varname: name, Value: value }

    return op, idx
}

func prsIdentExpr(idx int) (*ast.IdentExpr, int) {
    tokens := token.GetTokens()
    return &ast.IdentExpr{ Ident: tokens[idx] }, idx
}

func prsLitExpr(idx int) (*ast.LitExpr, int) {
    tokens := token.GetTokens()
    return &ast.LitExpr{ Val: tokens[idx], Type: types.TypeOfVal(tokens[idx].Str) }, idx
}

func prsUnaryExpr(idx int) (*ast.UnaryExpr, int) {
    tokens := token.GetTokens()
    expr := ast.UnaryExpr{ Operator: tokens[idx] }

    expr.Operand, idx = prsExpr(idx+1)

    return &expr, idx
}

func prsExpr(idx int) (ast.OpExpr, int) {
    tokens := token.GetTokens()
    value := tokens[idx]

    switch value.Type {
    case token.Plus, token.Minus:
        return prsUnaryExpr(idx)

    case token.Name:
        return prsIdentExpr(idx)

    case token.Number, token.Str:
        var expr ast.OpExpr
        expr, idx = prsLitExpr(idx)
        return expr, idx

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] no valid expression\n")
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
        return &ast.LitExpr{}, -1
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

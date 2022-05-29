package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

func isDec(idx int) bool {
    tokens := token.GetTokens()
    return tokens[idx+1].Type == token.Name && tokens[idx+2].Type == token.Typename
}

func prsDecl(idx int) (ast.OpDecl, int) {
    tokens := token.GetTokens()

    switch tokens[idx].Type {
    case token.Dec_var:
        var op ast.OpDecVar
        op, idx = prsDecVar(idx)
        return &op, idx

    case token.Def_var:
        var op ast.OpDefVar
        op, idx = prsDefVar(idx)
        return &op, idx

    case token.Def_fn:
        var op ast.OpDefFn
        op, idx = prsDefFn(idx)
        return &op, idx

    case token.Name:
        if tokens[idx+1].Type == token.ParenL {
            fmt.Fprintln(os.Stderr, "[ERROR] function calls are not allowed in global scope")
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
        } else if tokens[idx+1].Type == token.Assign {
            fmt.Fprintln(os.Stderr, "[ERROR] assigning variables is not allowed in global scope (but defining)")
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] variable \"%s\" is not used\n", tokens[idx].Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
            os.Exit(1)
        }
        return &ast.BadDecl{}, idx

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", tokens[idx].Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)

        return &ast.BadDecl{}, idx
    }
}

func prsDefFn(idx int) (ast.OpDefFn, int) {
    tokens := token.GetTokens()

    if tokens[idx+1].Str == "main" {
        isMainDefined = true
    }

    op := ast.OpDefFn{ FnName: tokens[idx+1] }
    op.Args, idx = prsDecArgs(idx+2)
    op.Block, idx = prsBlock(idx+1)

    return op, idx
}

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

    op := ast.OpDefVar{ Varname: name, Value: value }

    return op, idx
}

func prsDecArgs(idx int) ([]ast.OpDecVar, int) {
    tokens := token.GetTokens()

    decs := []ast.OpDecVar{}

    if (tokens[idx].Type != token.ParenL) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got \"%s\"(%s)\n", tokens[idx].Str, tokens[idx].Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx].At())
        os.Exit(1)
    }

    if isDec(idx) {
        var dec ast.OpDecVar
        dec, idx = prsDecVar(idx)
        decs = append(decs, dec)

        for tokens[idx+1].Type == token.Comma {
            dec, idx = prsDecVar(idx+1)
            decs = append(decs, dec)
        }
    }

    if (tokens[idx+1].Type != token.ParenR) {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got \"%s\"(%s)\n", tokens[idx+1].Str, tokens[idx+1].Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + tokens[idx+1].At())
        os.Exit(1)
    }

    return decs, idx+1
}

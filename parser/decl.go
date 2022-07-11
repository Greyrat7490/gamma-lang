package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

func prsDecl() ast.Decl {
    switch t := token.Next(); t.Type {
    case token.Def_fn:
        d := prsDefFn()
        return &d

    case token.Name:
        // define var (type is given)
        if token.Peek().Type == token.Typename {
            d := prsDefVar()
            return &d
        }
        // define var (infer the type with the value)
        if token.Peek().Type == token.Def_var {
            d := prsDefVarInfer()
            return &d
        }

        if token.Peek().Type == token.ParenL {
            fmt.Fprintln(os.Stderr, "[ERROR] function calls are not allowed in global scope")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }
        if token.Peek().Type == token.Assign {
            fmt.Fprintln(os.Stderr, "[ERROR] assigning variables is not allowed in global scope (but defining)")
            fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
            os.Exit(1)
        }
        fmt.Fprintf(os.Stderr, "[ERROR] identifier \"%s\" is not used (maybe forgot to previde a type)\n", token.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadDecl{}

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", t.Str)
        fmt.Fprintln(os.Stderr, "\t" + t.At())
        os.Exit(1)

        return &ast.BadDecl{}
    }
}

func prsDefFn() ast.DefFn {
    pos := token.Cur().Pos
    name := token.Next()

    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if name.Str == "main" {
        isMainDefined = true
    }

    op := ast.DefFn{ Pos: pos, FnName: name }
    token.Next()
    op.Args = prsDecArgs()
    token.Next()
    op.Block = prsBlock()

    return op
}

func prsDecVar() ast.DecVar {
    name := token.Cur()
    vartype := token.Next()
    end := token.Cur().Pos

    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
    if vartype.Type != token.Typename {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Typename but got %v\n", vartype)
        fmt.Fprintln(os.Stderr, "\t" + vartype.At())
        os.Exit(1)
    }

    t := types.ToType(vartype.Str)
    if t == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", vartype.Str)
        fmt.Fprintln(os.Stderr, "\t" + vartype.At())
        os.Exit(1)
    }

    return ast.DecVar{ Name: name, Type: t, TypePos: end }
}

func prsDefVar() ast.DefVar {
    dec := prsDecVar()
    if token.Peek().Type == token.Def_var {
        if token.Next().Type != token.Def_var {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a \":=\" but got %v\n", token.Cur())
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }
        pos := token.Cur().Pos

        token.Next()
        return ast.DefVar{ Name: dec.Name, Type: dec.Type, ColPos: pos, Value: prsExpr() }
    }

    fmt.Fprintln(os.Stderr, "[ERROR] declaring without initializing is not allowed")
    fmt.Fprintln(os.Stderr, "\t" + dec.Name.At())
    os.Exit(1)
    return ast.DefVar{}
}

func prsDefVarInfer() ast.DefVar {
    if token.Cur().Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got \"%s\"(%v)\n", token.Cur().Type, token.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    name := token.Cur()

    if token.Next().Type != token.Def_var {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \":=\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    pos := token.Cur().Pos

    token.Next()
    val := prsExpr()
    return ast.DefVar{ Name: name, Type: nil, ColPos: pos, Value: val }
}

func prsDecArgs() []ast.DecVar {
    decs := []ast.DecVar{}

    if token.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    if token.Next().Type != token.ParenR {
        decs = append(decs, prsDecVar())

        for token.Next().Type == token.Comma {
            token.Next()
            decs = append(decs, prsDecVar())
        }
    }

    if token.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return decs
}

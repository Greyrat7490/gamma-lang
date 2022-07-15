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
    case token.Fn:
        d := prsDefFn()
        return &d

    case token.Name:
        // define var/const (type is given)
        if isVarDec() {
            dec := prsDecVar()

            token.Next()
            if token.Cur().Type == token.DefVar {
                d := prsDefVar(dec)
                return &d
            }
            if token.Cur().Type == token.DefConst {
                d := prsDefConst(dec)
                return &d
            }

            fmt.Fprintln(os.Stderr, "[ERROR] declaring without initializing is not allowed")
            fmt.Fprintln(os.Stderr, "\t" + dec.At())
            os.Exit(1)
        }

        // define var (infer the type with the value)
        if token.Peek().Type == token.DefVar {
            d := prsDefVarInfer()
            return &d
        }
        // define const (infer the type with the value)
        if token.Peek().Type == token.DefConst {
            d := prsDefConstInfer()
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
    isPtr := false

    name := token.Cur()
    t := token.Next()
    if t.Type == token.Mul {
        t = token.Next()
        isPtr = true
    }

    end := token.Cur().Pos

    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
    if t.Type != token.Typename {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", t.Str)
        fmt.Fprintln(os.Stderr, "\t" + t.At())
        os.Exit(1)
    }

    return ast.DecVar{ Name: name, Type: types.ToType(t.Str, isPtr), TypePos: end }
}

func prsDefVar(dec ast.DecVar) ast.DefVar {
    pos := token.Cur().Pos
    token.Next()
    return ast.DefVar{ Name: dec.Name, Type: dec.Type, ColPos: pos, Value: prsExpr() }
}

func prsDefConst(dec ast.DecVar) ast.DefConst {
    return ast.DefConst(prsDefVar(dec))
}

func prsDefVarInfer() ast.DefVar {
    if token.Cur().Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got \"%s\"(%v)\n", token.Cur().Type, token.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    name := token.Cur()
    pos := token.Next().Pos

    token.Next()
    val := prsExpr()
    return ast.DefVar{ Name: name, Type: nil, ColPos: pos, Value: val }
}

func prsDefConstInfer() ast.DefConst {
    return ast.DefConst(prsDefVarInfer())
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

func isVarDec() bool {
    return token.Cur().Type == token.Name &&
        (token.Peek().Type == token.Typename || (token.Peek().Type == token.Mul && token.Peek2().Type == token.Typename))
}

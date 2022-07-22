package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/func"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
)

func prsDecl() ast.Decl {
    switch t := token.Next(); t.Type {
    case token.Fn:
        d := prsDefFn()
        return &d

    case token.Name:
        return prsDefine()

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", t.Str)
        fmt.Fprintln(os.Stderr, "\t" + t.At())
        os.Exit(1)

        return &ast.BadDecl{}
    }
}

func prsType() types.Type {
    isPtr := false
    typename := token.Next()
    if typename.Type == token.Mul {
        typename = token.Next()
        isPtr = true
    }
    if typename.Type != token.Typename {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", typename.Str)
        fmt.Fprintln(os.Stderr, "\t" + typename.At())
        os.Exit(1)
    }

    return types.ToType(typename.Str, isPtr)
}

func prsDecVar() ast.DecVar {
    name := token.Cur()
    t := prsType()
    end := token.Cur().Pos

    v := vars.DecVar(name, t)
    return ast.DecVar{ Ident: ast.Ident{ Ident: name, V: v }, Type: t, TypePos: end }
}

func prsDefVar(name token.Token, t types.Type) ast.DefVar {
    v := vars.DecVar(name, t)

    pos := token.Cur().Pos
    token.Next()
    return ast.DefVar{ Ident: ast.Ident{ Ident: name, V: v }, Type: t, ColPos: pos, Value: prsExpr() }
}

func prsDefConst(name token.Token, t types.Type) ast.DefConst {
    c := vars.DecConst(name, t)

    pos := token.Cur().Pos
    token.Next()
    return ast.DefConst{ Ident: ast.Ident{ Ident: name, C: c }, Type: t, ColPos: pos, Value: prsExpr() }
}

func prsDefVarInfer() ast.DefVar {
    name := token.Cur()
    pos := token.Next().Pos
    token.Next()
    val := prsExpr()

    t := val.GetType()
    v := vars.DecVar(name, t)
    return ast.DefVar{ Ident: ast.Ident{ Ident: name, V: v }, Type: t, ColPos: pos, Value: val }
}

func prsDefConstInfer() ast.DefConst {
    name := token.Cur()
    pos := token.Next().Pos
    token.Next()
    val := prsExpr()

    t := val.GetType()
    c := vars.DecConst(name, t)
    return ast.DefConst{ Ident: ast.Ident{ Ident: name, C: c }, Type: t, ColPos: pos, Value: val }
}

func prsDefine() ast.Decl {
    // define var/const (type is given)
    if isDec() {
        name := token.Cur()
        t := prsType()

        token.Next()
        if token.Cur().Type == token.DefVar {
            d := prsDefVar(name, t)
            return &d
        }
        if token.Cur().Type == token.DefConst {
            d := prsDefConst(name, t)
            return &d
        }

        fmt.Fprintln(os.Stderr, "[ERROR] declaring without initializing is not allowed")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
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
}


func prsDefFn() ast.DefFn {
    vars.CreateScope()
    defer vars.EndScope()

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

    var args []types.Type
    for _,a := range op.Args {
        args = append(args, a.Type)
    }
    fn.Declare(name, args)

    token.Next()
    op.Block = prsBlock()

    return op
}

func prsDecArgs() []ast.DecVar {
    decs := []ast.DecVar{}

    if token.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    if token.Next().Type != token.ParenR {
        d := prsDecVar()
        decs = append(decs, d)

        for token.Next().Type == token.Comma {
            token.Next()
            d := prsDecVar()
            decs = append(decs, d)
        }
    }

    if token.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return decs
}

func isDec() bool {
    return token.Cur().Type == token.Name &&
        (token.Peek().Type == token.Typename || (token.Peek().Type == token.Mul && token.Peek2().Type == token.Typename))
}

package prs

import (
    "os"
    "fmt"
    "strconv"
    "gorec/ast"
    "gorec/ast/identObj"
    "gorec/token"
    "gorec/types"
)

func prsDecl() ast.Decl {
    switch t := token.Next(); t.Type {
    case token.Fn:
        d := prsDefFn()
        return &d

    case token.Name:
        d := prsDefine()
        if _,ok := d.(*ast.BadDecl); ok {
            fmt.Fprintln(os.Stderr, "[ERROR] declaring without initializing is not allowed")
            fmt.Fprintln(os.Stderr, "\t" + d.At())
            os.Exit(1)
        }
        return d

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", t.Str)
        fmt.Fprintln(os.Stderr, "\t" + t.At())
        os.Exit(1)

        return &ast.BadDecl{}
    }
}

func prsType() types.Type {
    typename := token.Cur()

    switch typename.Type {
    case token.Mul:
        typename = token.Next()

        if typename.Type != token.Typename {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", typename.Str)
            fmt.Fprintln(os.Stderr, "\t" + typename.At())
            os.Exit(1)
        }

        if baseType := types.ToBaseType(typename.Str); baseType != nil {
            return types.PtrType{ BaseType: baseType }
        } else {
            return nil
        }

    case token.BrackL:
        token.Next()
        expr := prsExpr()

        if token.Next().Type != token.BrackR {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v but got %v\n", token.BrackR, token.Cur())
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }

        typename = token.Next()
        if typename.Type != token.Typename {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", typename.Str)
            fmt.Fprintln(os.Stderr, "\t" + typename.At())
            os.Exit(1)
        }

        eval := expr.ConstEval()
        if eval.Type != token.Number {
            if eval.Type == token.Unknown {
                fmt.Fprintln(os.Stderr, "[ERROR] lenght of an array has to a const/eval at compile time")
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] lenght of an array has to a Number but got (%v)\n", eval.Type)
            }
            fmt.Fprintln(os.Stderr, "\t" + eval.At())
            os.Exit(1)
        }

        lenght,_ := strconv.ParseUint(eval.Str, 10, 64)
        return types.ArrType{ Ptr: types.PtrType{ BaseType: types.ToBaseType(typename.Str) }, Len: lenght }
    default:
        return types.ToBaseType(typename.Str)
    }
}

func prsDecVar() ast.DecVar {
    name := token.Cur()
    token.Next()
    t := prsType()
    end := token.Cur().Pos

    v := identObj.DecVar(name, t)
    return ast.DecVar{ V: v, Type: t, TypePos: end }
}

func isDec() bool {
    return token.Cur().Type == token.Name &&
        (token.Peek().Type == token.Typename ||
            token.Peek().Type == token.Mul || token.Peek().Type == token.BrackL)
}
func isDefInfer() bool {
    return token.Cur().Type == token.Name &&
        (token.Peek().Type == token.DefVar || token.Peek().Type == token.DefConst)
}

func prsDefVar(name token.Token, t types.Type) ast.DefVar {
    v := identObj.DecVar(name, t)

    pos := token.Cur().Pos
    token.Next()
    return ast.DefVar{ V: v, Type: t, ColPos: pos, Value: prsExpr() }
}

func prsDefConst(name token.Token, t types.Type) ast.DefConst {
    c := identObj.DecConst(name, t)

    pos := token.Cur().Pos
    token.Next()
    return ast.DefConst{ C: c, Type: t, ColPos: pos, Value: prsExpr() }
}

func prsDefVarInfer(name token.Token) ast.DefVar {
    pos := token.Cur().Pos
    token.Next()
    val := prsExpr()

    t := val.GetType()
    v := identObj.DecVar(name, t)
    return ast.DefVar{ V: v, Type: t, ColPos: pos, Value: val }
}

func prsDefConstInfer(name token.Token) ast.DefConst {
    pos := token.Cur().Pos
    token.Next()
    val := prsExpr()

    t := val.GetType()
    c := identObj.DecConst(name, t)
    return ast.DefConst{ C: c, Type: t, ColPos: pos, Value: val }
}

func prsDefine() ast.Decl {
    name := token.Cur()
    token.Next()
    t := prsType()

    if t == nil {       // infer the type with the value
        if token.Cur().Type == token.DefVar {
            d := prsDefVarInfer(name)
            return &d
        }
        if token.Cur().Type == token.DefConst {
            d := prsDefConstInfer(name)
            return &d
        }
    } else {            // type is given
        if token.Next().Type == token.DefVar {
            d := prsDefVar(name, t)
            return &d
        }
        if token.Cur().Type == token.DefConst {
            d := prsDefConst(name, t)
            return &d
        }
    }

    return &ast.BadDecl{}
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

    f := identObj.DecFunc(name)

    identObj.StartScope()
    token.Next()
    argDecs := prsDecArgs()

    var args []types.Type
    for _,a := range argDecs {
        args = append(args, a.Type)
    }
    f.SetArgs(args)

    token.Next()
    block := prsBlock()
    identObj.EndScope()

    f.SetFrameSize(identObj.GetFrameSize())

    return ast.DefFn{ Pos: pos, F: f, Args: argDecs, Block: block }
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

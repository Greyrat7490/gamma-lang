package prs

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/struct"
)

func prsDecl() ast.Decl {
    switch t := token.Next(); t.Type {
    case token.Fn:
        d := prsDefFn()
        return &d

    case token.Struct:
        d := prsStruct()
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
        }

    case token.BrackL:
        return prsArrType()

    case token.Name:
        if obj := identObj.Get(typename.Str); obj != nil {
            if strct,ok := obj.(*structDec.Struct); ok {
                return strct.GetType()
            }
        }

    default:
        return types.ToBaseType(typename.Str)
    }

    return nil
}

func prsArrType() types.ArrType {
    if token.Cur().Type != token.BrackL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %v but got %v\n", token.BrackL, token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    var lens []uint64
    for token.Cur().Type == token.BrackL {
        token.Next()
        eval := cmpTime.ConstEval(prsExpr())
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
        lens = append(lens, lenght)

        if token.Next().Type != token.BrackR {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v but got %v\n", token.BrackR, token.Cur())
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }

        token.Next()
    }

    typename := token.Cur()
    if typename.Type != token.Typename {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", typename.Str)
        fmt.Fprintln(os.Stderr, "\t" + typename.At())
        os.Exit(1)
    }

    return types.ArrType{ Ptr: types.PtrType{ BaseType: types.ToBaseType(typename.Str) }, Lens: lens }
}

func prsDecVar() ast.DecVar {
    name, t := prsNameType()
    end := token.Cur().Pos

    return ast.DecVar{ V: identObj.DecVar(name, t), Type: t, TypePos: end }
}

func prsNameType() (name token.Token, typ types.Type) {
    name = token.Cur()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Last().At())
        os.Exit(1)
    }

    token.Next()
    typ = prsType()
    if typ == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", token.Last().Str)
        fmt.Fprintln(os.Stderr, "\t" + token.Last().At())
        os.Exit(1)
    }

    return
}

func isDec() bool {
    if token.Cur().Type != token.Name {
        return false
    }

    return isNextType()
}
func isDefInfer() bool {
    return token.Cur().Type == token.Name && (token.Peek().Type == token.DefVar || token.Peek().Type == token.DefConst)
}
func isNextType() bool {
    token.SaveIdx()
    defer token.ResetIdx()

    switch token.Next().Type {
    case token.Mul:
        typename := token.Next()

        if typename.Type != token.Typename {
            return false
        }

        return types.ToBaseType(typename.Str) != nil

    case token.BrackL:
        token.Next()
        expr := prsExpr()

        if token.Next().Type != token.BrackR {
            return false
        }

        typename := token.Next()
        if typename.Type != token.Typename {
            return false
        }

        return cmpTime.ConstEval(expr).Type != token.Number

    case token.Name:
        if obj := identObj.Get(token.Cur().Str); obj != nil {
            _,ok := obj.(*structDec.Struct)
            return ok
        }
        return false

    default:
        return types.ToBaseType(token.Cur().Str) != nil
    }
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

func prsStruct() ast.DefStruct {
    pos := token.Cur().Pos

    name := token.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    braceLPos := token.Next().Pos
    fields := prsDecFields()
    braceRPos := token.Cur().Pos

    var types []types.Type
    var names []string
    for _,f := range fields {
        types = append(types, f.Type)
        names = append(names, f.V.GetName())
    }
    s := identObj.DecStruct(name, names, types)

    return ast.DefStruct{ S: s, Pos: pos, Name: name, Fields: fields, BraceLPos: braceLPos, BraceRPos: braceRPos }
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

    identObj.StartScope()
    token.Next()
    argNames, argTypes := prsDecArgs()

    var retType types.Type = nil
    if token.Peek().Type == token.Arrow {
        token.Next()
        token.Next()
        retType = prsType()
    }

    if token.Next().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    f := identObj.DecFunc(name, argTypes, retType)

    var argDecs []ast.DecVar
    offset := uint(8)
    for i,t := range argTypes {
        if types.IsBigStruct(t) {
            argDecs = append(argDecs, ast.DecVar{ Type: t, V: identObj.DecArgFromStack(argNames[i], t, offset) })
            offset += t.Size()
        }
    }

    regCount := uint(0)
    for i,t := range argTypes {
        if types.IsBigStruct(t) {
            continue
        }

        needed := types.RegCount(t)
        if regCount + needed > 6 {
            argDecs = append(argDecs, ast.DecVar{ Type: t, V: identObj.DecArgFromStack(argNames[i], t, offset) })
            offset += t.Size()
        } else {
            argDecs = append(argDecs, ast.DecVar{ Type: t, V: identObj.DecArg(argNames[i], t) })
            regCount += needed
        }
    }

    block := prsBlock()
    f.SetFrameSize(identObj.GetFrameSize())
    identObj.EndScope()

    return ast.DefFn{ Pos: pos, F: f, Args: argDecs, RetType: retType, Block: block }
}

func prsDecFields() (decs []ast.DecVar) {
    if token.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    if token.Next().Type != token.BraceR {
        decs = append(decs, prsDecVar())

        for token.Next().Type == token.Comma {
            token.Next()
            decs = append(decs, prsDecVar())
        }
    }

    if token.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return
}

func prsDecArgs() (names []token.Token, types []types.Type) {
    if token.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    if token.Next().Type != token.ParenR {
        name,t := prsNameType()
        names = append(names, name)
        types = append(types, t)

        for token.Next().Type == token.Comma {
            token.Next()
            name,t := prsNameType()
            names = append(names, name)
            types = append(types, t)
        }
    }

    if token.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", token.Cur())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return
}

package prs

import (
    "os"
    "fmt"
    "gamma/types"
    "gamma/token"
    "gamma/import"
    "gamma/cmpTime"
    "gamma/ast"
    "gamma/ast/identObj"
    "gamma/ast/identObj/struct"
)

func prsDecl(tokens *token.Tokens) ast.Decl {
    switch t := tokens.Next(); t.Type {
    case token.Import:
        if !tokens.IsFileStart() {
            fmt.Fprintln(os.Stderr, "[ERROR] importing is only allowed at the beginning of a file")
            fmt.Fprintln(os.Stderr, "\t" + t.At())
            os.Exit(1)
        }

        d := prsImport(tokens)
        return &d

    case token.Fn:
        d := prsDefFn(tokens)
        return &d

    case token.Struct:
        d := prsStruct(tokens)
        return &d

    case token.Name:
        d := prsDefine(tokens)
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

func prsType(tokens *token.Tokens) types.Type {
    switch tokens.Cur().Type {
    case token.Mul:
        tokens.Next()
        return types.PtrType{ BaseType: prsType(tokens) }

    case token.BrackL:
        return prsArrType(tokens)

    case token.Name:
        if obj := identObj.Get(tokens.Cur().Str); obj != nil {
            if strct,ok := obj.(*structDec.Struct); ok {
                return strct.GetType()
            }
        }

        fmt.Fprintf(os.Stderr, "[ERROR] unknown struct type \"%s\"\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return nil

    default:
        return types.ToBaseType(tokens.Cur().Str)
    }
}

func prsArrType(tokens *token.Tokens) types.ArrType {
    if tokens.Cur().Type != token.BrackL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected %v but got %v\n", token.BrackL, tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    var lens []uint64
    for tokens.Cur().Type == token.BrackL {
        pos := tokens.Next().Pos
        expr := prsExpr(tokens)

        if length,ok := cmpTime.ConstEvalUint(expr); ok {
            lens = append(lens, length)
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] length of an array has to a const/eval at compile time")
            fmt.Fprintln(os.Stderr, "\t" + pos.At())
            os.Exit(1)
        }


        if tokens.Next().Type != token.BrackR {
            fmt.Fprintf(os.Stderr, "[ERROR] expected %v but got %v\n", token.BrackR, tokens.Cur())
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

        tokens.Next()
    }

    return types.ArrType{ Ptr: types.PtrType{ BaseType: prsType(tokens) }, Lens: lens }
}

func prsDecVar(tokens *token.Tokens) ast.DecVar {
    name, t := prsNameType(tokens)
    end := tokens.Cur().Pos

    return ast.DecVar{ V: identObj.DecVar(name, t), Type: t, TypePos: end }
}

func prsNameType(tokens *token.Tokens) (name token.Token, typ types.Type) {
    name = tokens.Cur()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Last().At())
        os.Exit(1)
    }

    tokens.Next()
    typ = prsType(tokens)
    if typ == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", tokens.Last().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Last().At())
        os.Exit(1)
    }

    return
}

func isDec(tokens *token.Tokens) bool {
    if tokens.Cur().Type != token.Name {
        return false
    }

    return isNextType(tokens)
}
func isDefInfer(tokens *token.Tokens) bool {
    return tokens.Cur().Type == token.Name && (tokens.Peek().Type == token.DefVar || tokens.Peek().Type == token.DefConst)
}
func isNextType(tokens *token.Tokens) bool {
    tokens.SaveIdx()
    defer tokens.ResetIdx()

    switch tokens.Next().Type {
    case token.Mul:
        typename := tokens.Next()

        if typename.Type != token.Typename {
            return false
        }

        return types.ToBaseType(typename.Str) != nil

    case token.BrackL:
        tokens.Next()
        expr := prsExpr(tokens)

        if tokens.Next().Type != token.BrackR {
            return false
        }

        typename := tokens.Next()
        if typename.Type != token.Typename {
            return false
        }

        kind := cmpTime.ConstEval(expr).GetKind()
        return kind == types.Int || kind == types.Uint

    case token.Name:
        if obj := identObj.Get(tokens.Cur().Str); obj != nil {
            _,ok := obj.(*structDec.Struct)
            return ok
        }
        return false

    default:
        return types.ToBaseType(tokens.Cur().Str) != nil
    }
}

func prsDefVar(tokens *token.Tokens, name token.Token, t types.Type) ast.DefVar {
    v := identObj.DecVar(name, t)

    pos := tokens.Cur().Pos
    tokens.Next()
    val := prsExpr(tokens)
    return ast.DefVar{ V: v, Type: t, ColPos: pos, Value: val }
}

func prsDefConst(tokens *token.Tokens, name token.Token, t types.Type) ast.DefConst {
    pos := tokens.Cur().Pos
    tokens.Next()
    val := prsExpr(tokens)

    v := cmpTime.ConstEval(val)
    if v == nil {
        fmt.Fprintln(os.Stderr, "[ERROR] expected a const expr")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    return ast.DefConst{ C: identObj.DecConst(name, t, v), Type: t, ColPos: pos, Value: val }
}

func prsDefVarInfer(tokens *token.Tokens, name token.Token) ast.DefVar {
    pos := tokens.Cur().Pos
    tokens.Next()
    val := prsExpr(tokens)

    t := val.GetType()
    if t == nil {
        if f,ok := val.(*ast.FnCall); ok {
            fmt.Fprintf(os.Stderr, "[ERROR] %s returns nothing\n", f.Ident.Name)
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] could not get Type of the expr")
        }
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }
    v := identObj.DecVar(name, t)
    return ast.DefVar{ V: v, Type: t, ColPos: pos, Value: val }
}

func prsDefConstInfer(tokens *token.Tokens, name token.Token) ast.DefConst {
    pos := tokens.Cur().Pos
    tokens.Next()
    val := prsExpr(tokens)

    t := val.GetType()
    if t == nil {
        if f,ok := val.(*ast.FnCall); ok {
            fmt.Fprintf(os.Stderr, "[ERROR] %s returns nothing\n", f.Ident.Name)
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] could not get Type of the expr")
        }
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }
    v := cmpTime.ConstEval(val)
    if v == nil {
        fmt.Fprintln(os.Stderr, "[ERROR] expected a const expr")
        fmt.Fprintln(os.Stderr, "\t" + val.At())
        os.Exit(1)
    }

    return ast.DefConst{ C: identObj.DecConst(name, t, v), Type: t, ColPos: pos, Value: val }
}

func prsDefine(tokens *token.Tokens) ast.Decl {
    name := tokens.Cur()
    tokens.Next()
    t := prsType(tokens)

    if t == nil {       // infer the type with the value
        if tokens.Cur().Type == token.DefVar {
            d := prsDefVarInfer(tokens, name)
            return &d
        }
        if tokens.Cur().Type == token.DefConst {
            d := prsDefConstInfer(tokens, name)
            return &d
        }
    } else {            // type is given
        if tokens.Next().Type == token.DefVar {
            d := prsDefVar(tokens, name, t)
            return &d
        }
        if tokens.Cur().Type == token.DefConst {
            d := prsDefConst(tokens, name, t)
            return &d
        }
    }

    return &ast.BadDecl{}
}

func prsStruct(tokens *token.Tokens) ast.DefStruct {
    pos := tokens.Cur().Pos

    name := tokens.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    braceLPos := tokens.Next().Pos
    fields := prsDecFields(tokens)
    braceRPos := tokens.Cur().Pos

    var names []string
    var types []types.Type
    for _,f := range fields {
        names = append(names, f.Name.Str)
        types = append(types, f.Type)
    }

    return ast.DefStruct{ S: identObj.DecStruct(name, names, types), Pos: pos, Name: name, BraceLPos: braceLPos, Fields: fields, BraceRPos: braceRPos }
}

func prsDefFn(tokens *token.Tokens) ast.DefFn {
    pos := tokens.Cur().Pos
    name := tokens.Next()

    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
    if name.Str == "main" {
        isMainDefined = true
    }

    identObj.StartScope()
    tokens.Next()
    argNames, argTypes := prsDecArgs(tokens)

    var retType types.Type = nil
    if tokens.Peek().Type == token.Arrow {
        tokens.Next()
        tokens.Next()
        retType = prsType(tokens)
    }

    if tokens.Next().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
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

    block := prsBlock(tokens)
    f.SetFrameSize(identObj.GetFrameSize())
    identObj.EndScope()

    def := ast.DefFn{ Pos: pos, F: f, Args: argDecs, RetType: retType, Block: block }
    cmpTime.AddConstFunc(def)
    return def
}

func prsDecField(tokens *token.Tokens) ast.DecField {
    name,t := prsNameType(tokens)
    return ast.DecField{ Name: name, Type: t, TypePos: tokens.Cur().Pos }
}

func prsDecFields(tokens *token.Tokens) (fields []ast.DecField) {
    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if tokens.Next().Type != token.ParenR {
        fields = append(fields, prsDecField(tokens))

        for tokens.Next().Type == token.Comma {
            tokens.Next()
            fields = append(fields, prsDecField(tokens))
        }
    }

    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return
}

func prsDecArgs(tokens *token.Tokens) (names []token.Token, types []types.Type) {
    if tokens.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if tokens.Next().Type != token.ParenR {
        name,t := prsNameType(tokens)
        names = append(names, name)
        types = append(types, t)

        for tokens.Next().Type == token.Comma {
            tokens.Next()
            name,t := prsNameType(tokens)
            names = append(names, name)
            types = append(types, t)
        }
    }

    if tokens.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return
}

func prsImport(tokens *token.Tokens) ast.Import {
    pos := tokens.Cur().Pos
    path := tokens.Next()

    if path.Type != token.Str {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a path as string but got %v\n", path)
        fmt.Fprintln(os.Stderr, "\t" + path.At())
        os.Exit(1)
    }

    d := ast.Import{ Pos: pos, Path: path }

    if tokens, isNew := imprt.Import(path); isNew {
        for tokens.Peek().Type != token.EOF {
            tokens.SetLastImport()
            d.Decls = append(d.Decls, prsDecl(tokens))
        }

        imprt.EndImport(tokens.GetPath())
    }

    return d
}

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
        d := prsDefFn(tokens, false, false)
        return &d

    case token.ConstFn:
        d := prsDefFn(tokens, true, false)
        return &d

    case token.Struct:
        d := prsStruct(tokens)
        return &d

    case token.Interface:
        return prsInterface(tokens)

    case token.Impl:
        return prsImpl(tokens)

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
        if tokens.Peek().Type == token.XSwitch {
            return prsVecType(tokens)
        } else {
            return prsArrType(tokens)
        }

    case token.Name:
        if obj := identObj.Get(tokens.Cur().Str); obj != nil {
            if strct,ok := obj.(*identObj.Struct); ok {
                return strct.GetType()
            }
        }

        if generic := identObj.GetGeneric(tokens.Cur().Str); generic != nil {
            return generic
        }

        fmt.Fprintf(os.Stderr, "[ERROR] unknown struct type \"%s\"\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return nil

    default:
        t := types.ToBaseType(tokens.Cur().Str)
        if t == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] %s is not a valid type\n", tokens.Cur().Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }
        return t
    }
}

func prsVecType(tokens *token.Tokens) types.VecType {
    if tokens.Cur().Type != token.BrackL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"[\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    if tokens.Next().Type != token.XSwitch {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"$\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    if tokens.Next().Type != token.BrackR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"]\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    tokens.Next()
    return types.VecType{ BaseType: prsType(tokens) }
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

    return types.ArrType{ BaseType: prsType(tokens), Lens: lens }
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

    return
}

func isDec(tokens *token.Tokens) bool {
    if tokens.Cur().Type != token.Name {
        return false
    }

    return isNextType(tokens)
}
func isDefInfer(tokens *token.Tokens) bool {
    if tokens.Cur().Type != token.Name {
        return false
    }

    if tokens.Peek().Type == token.DefVar {
        return true
    }

    if tokens.Peek().Type == token.DefConst {
        if tokens.Peek2().Type == token.Lss || isStruct(tokens.Cur()) {
            return false
        }
        return true
    }

    return false
}
func isStruct(token token.Token) bool {
    _,ok := identObj.Get(token.Str).(*identObj.Struct)
    return ok
}
func isGenericFunc(token token.Token) bool {
    if f,ok := identObj.Get(token.Str).(*identObj.Func); ok {
        return f.GetGeneric() != nil
    }

    return false
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
            if _,ok := obj.(*identObj.Struct); ok {
                return true
            }
        }

        if generic := identObj.GetGeneric(tokens.Cur().Str); generic != nil {
            return true
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

    var t types.Type = nil
    tokens.Next()
    if tokens.Cur().Type != token.DefVar && tokens.Cur().Type != token.DefConst {
        t = prsType(tokens)
    }

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

    return ast.DefStruct{ S: identObj.DecStruct(name, names, types), Pos: pos, 
        Name: name, BraceLPos: braceLPos, Fields: fields, BraceRPos: braceRPos }
}

func prsInterface(tokens *token.Tokens) ast.Decl {
    pos := tokens.Cur().Pos

    name := tokens.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    braceLPos := tokens.Next().Pos
    identObj.StartScope()
    I := identObj.DecInterface(name)

    heads := make([]ast.FnHead, 0)
    funcs := make([]identObj.Func, 0)

    for tokens.Next().Type != token.BraceR {
        identObj.StartScope()
        fnHead := prsFnHead(tokens, false, true)
        identObj.EndScope()

        heads = append(heads, fnHead)
        funcs = append(funcs, *fnHead.F)
    }

    I.Funcs = funcs

    braceRPos := tokens.Cur().Pos
    identObj.EndScope()

    return &ast.DefInterface{ Pos: pos, Name: name, I: I, BraceLPos: braceLPos, BraceRPos: braceRPos, FnHeads: heads }
}

func prsImpl(tokens *token.Tokens) ast.Decl {
    pos := tokens.Cur().Pos

    StructName := tokens.Next()
    if StructName.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", StructName)
        fmt.Fprintln(os.Stderr, "\t" + StructName.At())
        os.Exit(1)
    }

    // TODO: no interface given -> create new one for the struct
    if tokens.Next().Type != token.DefConst {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \"::\" but got %v\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    InterfaceName := tokens.Next()
    if InterfaceName.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", InterfaceName)
        fmt.Fprintln(os.Stderr, "\t" + InterfaceName.At())
        os.Exit(1)
    }

    S,ok := identObj.Get(StructName.Str).(*identObj.Struct)
    if !ok || S == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] struct \"%s\" is not defined\n", StructName.Str)
        fmt.Fprintln(os.Stderr, "\t" + StructName.At())
        os.Exit(1)
    }

    sType := S.GetType().(types.StructType)
    identObj.CurImplStruct = &sType

    I,ok := identObj.Get(InterfaceName.Str).(*identObj.Interface)
    if !ok || I == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] interface \"%s\" is not defined\n", InterfaceName.Str)
        fmt.Fprintln(os.Stderr, "\t" + InterfaceName.At())
        os.Exit(1)
    }

    braceLPos := tokens.Next().Pos
    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \"{\" but got %v\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    identObj.StartScope()

    impl := identObj.CreateImpl(pos, I, S)
    S.AddImpl(impl)

    funcs := make([]ast.DefFn, 0, len(I.Funcs))

    for tokens.Next().Type != token.BraceR {
        switch tokens.Cur().Type {
        case token.Fn:
            funcs = append(funcs, prsDefFn(tokens, false, true))
        case token.ConstFn:
            funcs = append(funcs, prsDefFn(tokens, true, true))
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] you can only define methods in impl (unexpected token %v)\n", tokens.Cur().Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }
    }

    braceRPos := tokens.Cur().Pos
    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \"}\" but got %v\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    identObj.EndScope()
    identObj.CurImplStruct = nil

    return &ast.Impl{ Pos: pos, Impl: impl, BraceLPos: braceLPos, BraceRPos: braceRPos, FnDefs: funcs }
}

func prsFnHead(tokens *token.Tokens, isConst bool, isMethod bool) ast.FnHead {
    name := tokens.Cur()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    var f *identObj.Func = nil
    if isMethod && identObj.CurImplStruct != nil {
        f = identObj.DecMethod(name, isConst, identObj.CurImplStruct.Name)
    } else {
        f = identObj.DecFunc(name, isConst)
    }

    tokens.Next()
    generic := prsGeneric(tokens)
    isGeneric := generic.Str != ""
    if isGeneric {
        f.SetGeneric(&types.GenericType{ Name: generic.Str, UsedTypes: make([]types.Type, 0) })
    }

    recver, argNames, argTypes := prsArgs(tokens, isMethod)
    f.SetArgs(recver, argTypes)

    if name.Str == "main" {
        isMainDefined = true
        noMainArg = len(argNames) == 0
    }

    var retType types.Type = nil
    if tokens.Peek().Type == token.Arrow {
        tokens.Next()
        tokens.Next()
        retType = prsType(tokens)
    }
    f.SetRetType(retType)

    var recverDec *ast.DecVar = nil
    if recver != nil && identObj.CurImplStruct != nil {
        name := token.Token{ Str: "self", Pos: recver.DecPos, Type: token.Name }
        var t types.Type = nil
        if recver.IsPtr {
            t = &types.PtrType{ BaseType: *identObj.CurImplStruct }
        } else {
            t = *identObj.CurImplStruct
        }

        recverDec = &ast.DecVar{ Type: t, V: identObj.DecVar(name, t) }
    }

    var argDecs []ast.DecVar
    for i,t := range argTypes {
        argDecs = append(argDecs, ast.DecVar{ Type: t, V: identObj.DecVar(argNames[i], t) })
    }

    return ast.FnHead{ Name: name, F: f, Recver: recverDec, Args: argDecs, RetType: retType, IsConst: isConst, IsGeneric: isGeneric, Generic: generic }
}

func prsDefFn(tokens *token.Tokens, isConst bool, isMethod bool) ast.DefFn {
    pos := tokens.Cur().Pos
    tokens.Next()

    identObj.StartScope()
    fnHead := prsFnHead(tokens, isConst, isMethod)

    if tokens.Next().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    identObj.StartScope()
    block := prsBlock(tokens)
    identObj.EndScope()

    identObj.EndScope()

    def := ast.DefFn{ Pos: pos, FnHead: fnHead, Block: block }

    if isConst {
        cmpTime.AddConstFunc(def)
    }

    return def
}

func prsGeneric(tokens *token.Tokens) token.Token {
    if tokens.Cur().Type == token.Lss {
        name := tokens.Next()

        if name.Type != token.Name {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", tokens.Cur())
            fmt.Fprintln(os.Stderr, "\t" + tokens.Last().At())
            os.Exit(1)
        }

        if tokens.Next().Type != token.Grt {
            fmt.Fprintf(os.Stderr, "[ERROR] expected \">\" but got %v\n", tokens.Cur())
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

        tokens.Next()
        return name
    }

    return token.Token{}
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
    if tokens.Peek().Type == token.BraceR { tokens.Next(); return }    // empty struct

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

func prsRecver(tokens *token.Tokens) *identObj.FnRecver {
    if tokens.Cur().Type == token.Mul && tokens.Peek().Type == token.Self {
        tokens.Next()
        return &identObj.FnRecver{ DecPos: tokens.Cur().Pos, IsPtr: true }
    }

    if tokens.Cur().Type == token.Self {
        return &identObj.FnRecver{ DecPos: tokens.Cur().Pos, IsPtr: false }
    }

    return nil
}

func prsArgs(tokens *token.Tokens, isMethod bool) (recver *identObj.FnRecver, names []token.Token, types []types.Type) {
    if tokens.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if tokens.Next().Type != token.ParenR {
        recver = prsRecver(tokens)

        if recver == nil {
            name,t := prsNameType(tokens)
            names = append(names, name)
            types = append(types, t)
        } else if !isMethod {
            fmt.Fprintln(os.Stderr, "[ERROR] Self can only be used for methods (inside impl)")
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

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

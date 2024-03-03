package prs

import (
	"fmt"
	"gamma/ast"
	"gamma/ast/identObj"
	"gamma/cmpTime"
	"gamma/import"
	"gamma/token"
	"gamma/types"
	"os"
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

    case token.Fn, token.ConstFn:
        d := prsDefFn(tokens, false)
        return &d

    case token.Struct:
        d := prsStruct(tokens)
        return &d

    case token.Enum:
        return prsEnum(tokens)

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

func createSelfType(tokens *token.Tokens) types.Type {
    if identObj.CurSelfType == nil {
        fmt.Fprintln(os.Stderr, "[ERROR] Self used outside of impl and interface")
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
        return nil
    }

    return identObj.CurSelfType
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

    pos := tokens.Next().Pos
    expr := prsExpr(tokens)

    Len := uint64(0)
    if length,ok := cmpTime.ConstEvalUint(expr); ok {
        Len = length
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

    return types.ArrType{ BaseType: prsType(tokens), Len: Len }
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
    if tokens.Cur().Type != token.Name && tokens.Cur().Type != token.UndScr {
        return false
    }

    if tokens.Peek().Type == token.DefVar {
        return true
    }

    if tokens.Peek().Type == token.DefConst {
        if tokens.Peek2().Type == token.Lss || isImplementable(tokens.Cur()) || isInterface(tokens.Cur()) {
            return false
        }
        return true
    }

    return false
}
func isStruct(obj identObj.IdentObj) bool {
    _,ok := obj.(*identObj.Struct)
    return ok
}
func isEnum(token token.Token) bool {
    _,ok := identObj.Get(token.Str).(*identObj.Enum)
    return ok
}
func isImplementable(token token.Token) bool {
    return identObj.GetImplObj(token.Str) != nil
}
func isInterface(token token.Token) bool {
    _,ok := identObj.Get(token.Str).(*identObj.Interface)
    return ok
}
func isEnumLit(enumType types.EnumType, elemName string) bool {
    return enumType.HasElem(elemName)
}
func isGenericFunc(name string) bool {
    if f,ok := identObj.Get(name).(*identObj.Func); ok {
        return f.IsGeneric()
    }

    return false
}

func isType(tokens *token.Tokens) bool {
    savedIdx := tokens.SaveIdx()
    defer tokens.ResetIdx(savedIdx)
    return isType_(tokens)
}
func isType_(tokens *token.Tokens) bool {
    switch tokens.Cur().Type {
    case token.Typename, token.SelfType:
        return true

    case token.Mul:
        tokens.Next()
        return isType_(tokens)

    case token.BrackL:
        tokens.Next()

        if tokens.Cur().Type != token.XSwitch {
            idxKind := prsExpr(tokens).GetType().GetKind()
            if idxKind != types.Int && idxKind != types.Uint && idxKind != types.Infer {
                return false
            }
        }

        if tokens.Next().Type != token.BrackR {
            return false
        }

        tokens.Next()
        return isType_(tokens)

    case token.Name:
        if obj := identObj.Get(tokens.Cur().Str); obj != nil {
            if _,ok := obj.(*identObj.Struct); ok {
                return true
            }

            if _,ok := obj.(*identObj.Interface); ok {
                return true
            }

            if _,ok := obj.(*identObj.Enum); ok {
                return true
            }

            if _,ok := obj.(*identObj.Generic); ok {
                return true
            }
        }

        return false

    default:
        return false
    }
}
func isNextType(tokens *token.Tokens) bool {
    savedIdx := tokens.SaveIdx()
    defer tokens.ResetIdx(savedIdx)
    tokens.Next()
    return isType_(tokens)
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

    case token.SelfType:
        return createSelfType(tokens)

    case token.Name:
        if obj := identObj.Get(tokens.Cur().Str); obj != nil {
            var t types.Type = nil
            if strct,ok := obj.(*identObj.Struct); ok {
                t = strct.GetType()
            }

            if interfc,ok := obj.(*identObj.Interface); ok {
                t = interfc.GetType()
            }

            if enum,ok := obj.(*identObj.Enum); ok {
                t = enum.GetType()
            }

            if gen,ok := obj.(*identObj.Generic); ok {
                t = gen.GetType()
            }

            if tokens.Peek().Type == token.Lss {
                tokens.Next()
                tokens.Next()
                t = types.ReplaceGeneric(t, prsType(tokens))
                if tokens.Next().Type != token.Grt {
                    fmt.Fprintf(os.Stderr, "[ERROR] expected \">\" but got %s\n", tokens.Cur().Str)
                    fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                    os.Exit(1)
                }
            }

            return t
        }

        fmt.Fprintf(os.Stderr, "[ERROR] type \"%s\" is not defined\n", tokens.Cur().Str)
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

func prsInterfaceType(tokens *token.Tokens) *types.InterfaceType {
    if obj := identObj.Get(tokens.Cur().Str); obj != nil {
        if interfc,ok := obj.(*identObj.Interface); ok {
            t := interfc.GetType().(types.InterfaceType)

            if tokens.Peek().Type == token.Lss {
                tokens.Next()
                tokens.Next()
                t.Generic.SetType = prsType(tokens)
                if tokens.Next().Type != token.Grt {
                    fmt.Fprintf(os.Stderr, "[ERROR] expected \">\" but got %s\n", tokens.Cur().Str)
                    fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
                    os.Exit(1)
                }
            }

            return &t
        }
    }

    return nil
}

func prsDefVar(tokens *token.Tokens, name token.Token, t types.Type) ast.DefVar {
    pos := tokens.Cur().Pos
    tokens.Next()
    val := prsExpr(tokens)

    v := identObj.DecVar(name, t)
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
    v := identObj.DecVar(name, t)
    return ast.DefVar{ V: v, Type: t, ColPos: pos, Value: val }
}

func prsDefConstInfer(tokens *token.Tokens, name token.Token) ast.DefConst {
    pos := tokens.Cur().Pos
    tokens.Next()
    val := prsExpr(tokens)

    t := val.GetType()
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
    if isNextType(tokens) {
        tokens.Next()
        t = prsType(tokens)
    }

    tokens.Next()

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
        if tokens.Cur().Type == token.DefVar {
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

    if !identObj.InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare a struct in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    identObj.StartScope()
    defer identObj.EndScope()

    name := tokens.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    tokens.Next()
    generic := prsGeneric(tokens)

    s := identObj.DecStruct(name, generic)

    braceLPos := tokens.Cur().Pos
    fields := prsDecFields(tokens)
    braceRPos := tokens.Cur().Pos

    var names []string
    var ts []types.Type
    for _,f := range fields {
        names = append(names, f.Name.Str)
        ts = append(ts, f.Type)
    }
    s.SetFields(names, ts)

    return ast.DefStruct{ S: s, Pos: pos, Name: name, BraceLPos: braceLPos, Fields: fields, BraceRPos: braceRPos }
}

func prsInterface(tokens *token.Tokens) ast.Decl {
    pos := tokens.Cur().Pos

    if !identObj.InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare an interface in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    identObj.StartScope()
    defer identObj.EndScope()

    tokens.Next()
    generic := prsGeneric(tokens)

    name := tokens.Cur()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    braceLPos := tokens.Next().Pos
    I := identObj.DecInterface(name, generic)
    identObj.CurSelfType = I.GetType()

    heads := make([]ast.FnHead, 0)

    for tokens.Next().Type != token.BraceR {
        identObj.StartScope()
        fnHead := prsFnHead(tokens, true)
        identObj.EndScope()

        heads = append(heads, fnHead)
        I.AddFunc(fnHead.F)
    }

    braceRPos := tokens.Cur().Pos
    identObj.CurSelfType = nil

    return &ast.DefInterface{ Pos: pos, Name: name, I: I, BraceLPos: braceLPos, BraceRPos: braceRPos, FnHeads: heads }
}

func prsEnum(tokens *token.Tokens) ast.Decl {
    pos := tokens.Cur().Pos

    if !identObj.InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare an enum in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    identObj.StartScope()
    defer identObj.EndScope()

    name := tokens.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    tokens.Next()
    generic := prsGeneric(tokens)

    var idTyp types.Type = nil
    if tokens.Cur().Type != token.BraceL {
        idTyp = prsType(tokens)
        tokens.Next()
    } else {
        idTyp = types.CreateUint(types.Ptr_Size)
    }

    e := identObj.DecEnum(name, generic)

    braceLPos := tokens.Cur().Pos
    elems := prsEnumElems(tokens)
    braceRPos := tokens.Cur().Pos

    var names []string
    var ts []types.Type
    for _,e := range elems {
        names = append(names, e.Name.Str)
        if e.Type == nil {
            ts = append(ts, nil)
        } else {
            ts = append(ts, e.Type.Type)
        }
    }
    e.SetElems(idTyp, names, ts)

    return &ast.DefEnum{ E: e, Pos: pos, IdType: idTyp, Name: name, BraceLPos: braceRPos, Elems: elems, BraceRPos: braceLPos }
}
func prsEnumElems(tokens *token.Tokens) []ast.EnumElem {
    res := make([]ast.EnumElem, 0, 3)

    if tokens.Next().Type != token.BraceR {
        res = append(res, prsEnumElem(tokens))

        for tokens.Next().Type == token.Comma {
            tokens.Next()
            res = append(res, prsEnumElem(tokens))
        }
    }

    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    return res
}
func prsEnumElem(tokens *token.Tokens) ast.EnumElem {
    name := tokens.Cur()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    t := prsEnumElemType(tokens)

    return ast.EnumElem{ Name: name, Type: t }
}
func prsEnumElemType(tokens *token.Tokens) *ast.EnumElemType {
    if tokens.Peek().Type == token.ParenL {
        posL := tokens.Next().Pos
        tokens.Next()
        typ := prsType(tokens)
        posR := tokens.Next().Pos
        return &ast.EnumElemType{ ParenL: posL, Type: typ, ParenR: posR }
    }

    return nil
}

func prsImpl(tokens *token.Tokens) ast.Decl {
    pos := tokens.Cur().Pos

    if !identObj.InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare an impl in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    identObj.StartScope()
    defer identObj.EndScope()

    tokens.Next()
    generic := prsGeneric(tokens)
    dstType := prsType(tokens)

    var interfaceType *types.InterfaceType = nil
    if tokens.Peek().Type == token.DefConst {
        tokens.Next()
        tokens.Next()
        interfaceType = prsInterfaceType(tokens)
    }

    implObj := identObj.CreateImplObj(dstType)

    identObj.CurSelfType = dstType

    braceLPos := tokens.Next().Pos
    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \"{\" but got %v\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    impl := identObj.CreateImpl(pos, interfaceType, dstType, generic)
    implObj.AddImpl(impl)

    funcsReservedLen := 5
    if interfaceType != nil { funcsReservedLen = len(interfaceType.Funcs) }
    funcs := make([]ast.DefFn, 0, funcsReservedLen)

    for tokens.Next().Type != token.BraceR {
        if tokens.Cur().Type != token.Fn && tokens.Cur().Type != token.ConstFn {
            fmt.Fprintf(os.Stderr, "[ERROR] you can only define funcs in impl (unexpected token %v)\n", tokens.Cur().Str)
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

        funcs = append(funcs, prsDefFn(tokens, true))
    }

    braceRPos := tokens.Cur().Pos
    if tokens.Cur().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \"}\" but got %v\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    identObj.CurSelfType = nil

    return &ast.Impl{ Pos: pos, Impl: impl, BraceLPos: braceLPos, BraceRPos: braceRPos, FnDefs: funcs }
}

func prsFnHead(tokens *token.Tokens, isInterfaceFn bool) ast.FnHead {
    fn := tokens.Cur()
    if fn.Type != token.Fn && fn.Type != token.ConstFn {
        fmt.Fprintf(os.Stderr, "[ERROR] expected fn or cfn but got %v\n", fn.Str)
        fmt.Fprintln(os.Stderr, "\t" + fn.At())
        os.Exit(1)
    }

    isConst := fn.Type == token.ConstFn

    name := tokens.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    tokens.Next()
    generic := prsGeneric(tokens)
    f := identObj.DecFunc(name, isConst, identObj.CurSelfType, generic)

    argNames, argTypes := prsArgs(tokens, isInterfaceFn)
    f.SetArgs(argTypes)

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

    var argDecs []ast.DecVar
    for i,t := range argTypes {
        a := ast.DecVar{ Type: t, V: identObj.DecVar(argNames[i], t) }
        argDecs = append(argDecs, a)
    }

    return ast.FnHead{ Name: name, F: f, Generic: generic, Args: argDecs, RetType: retType, IsConst: isConst }
}

func prsDefFn(tokens *token.Tokens, isInterfaceFn bool) ast.DefFn {
    pos := tokens.Cur().Pos

    identObj.StartScope()
    fnHead := prsFnHead(tokens, isInterfaceFn)

    if tokens.Next().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    block := prsBlock(tokens)

    identObj.EndScope()

    def := ast.DefFn{ Pos: pos, FnHead: fnHead, Block: block }

    if fnHead.IsConst {
        cmpTime.AddConstFunc(def)
    }

    return def
}

func prsGeneric(tokens *token.Tokens) *identObj.Generic {
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
        return identObj.DecGeneric(name)
    }

    return nil
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

func prsOptionalSelfType(tokens *token.Tokens) types.Type {
    if tokens.Peek().Type == token.SelfType {
        tokens.Next()
        return createSelfType(tokens)
    }

    if tokens.Peek().Type == token.Mul && tokens.Peek2().Type == token.SelfType {
        tokens.Next()
        tokens.Next()
        return types.PtrType{ BaseType: createSelfType(tokens) } 
    }

    // explicitly used StructName
    if identObj.CurSelfType != nil {
        if tokens.Peek().Type == token.Name && tokens.Peek().Str == identObj.CurSelfType.String() {
            tokens.Next()
            return identObj.CurSelfType

        } else if tokens.Peek().Type == token.Mul && tokens.Peek2().Type == token.Name && tokens.Peek2().Str == identObj.CurSelfType.String() {
            tokens.Next()
            tokens.Next()
            return types.PtrType{ BaseType: identObj.CurSelfType } 
        }
    }

    return nil
}

func prsSelf(tokens *token.Tokens) (name token.Token, typ types.Type) {
    if tokens.Cur().Type == token.Mul && tokens.Peek().Type == token.Self {
        return tokens.Next(), types.PtrType{ BaseType: createSelfType(tokens) } 
    }

    if tokens.Cur().Type == token.Self {
        name = tokens.Cur()
        typ = prsOptionalSelfType(tokens)
        if typ == nil {
            typ = createSelfType(tokens)
        }
        return
    }

    name = tokens.Cur()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    typ = prsOptionalSelfType(tokens)

    return
}

func prsArgs(tokens *token.Tokens, isInterfaceFn bool) (names []token.Token, types []types.Type) {
    if tokens.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got %v\n", tokens.Cur())
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }

    if tokens.Next().Type != token.ParenR {
        name, t := prsSelf(tokens)

        if t == nil {
            name,t = prsNameType(tokens)
        } else if !isInterfaceFn {
            fmt.Fprintln(os.Stderr, "[ERROR] Self can only be used for interface funcs (inside interface / impl)")
            fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
            os.Exit(1)
        }

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

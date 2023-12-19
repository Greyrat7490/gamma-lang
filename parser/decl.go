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

func createSelfType() types.Type {
    if identObj.CurSelfType != nil {
        return identObj.CurSelfType
    }

    return types.StructType{ Name: "Self" }
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
        return createSelfType()

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

        if generic := identObj.GetGeneric(tokens.Cur().Str); generic != nil {
            return generic
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
        if tokens.Peek2().Type == token.Lss || isImplementable(tokens.Cur()) {
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
    _,ok := identObj.Get(token.Str).(identObj.Implementable)
    return ok
}
func isEnumLit(enumName string, elemName string) bool {
    if e,ok := identObj.Get(enumName).(*identObj.Enum); ok {
        return e.GetFunc(elemName) == nil
    }

    return false
}
func isGenericFunc(name string) bool {
    if f,ok := identObj.Get(name).(*identObj.Func); ok {
        return f.GetGeneric() != nil
    }

    return false
}
func isNextType_(tokens *token.Tokens) bool {
    switch tokens.Next().Type {
    case token.Mul:
        return isNextType_(tokens)

    case token.BrackL:
        tokens.Next()
        idxKind := prsExpr(tokens).GetType().GetKind()
        if idxKind != types.Int && idxKind != types.Uint {
            return false
        }

        if tokens.Next().Type != token.BrackR {
            return false
        }

        return isNextType_(tokens)

    case token.Name:
        if obj := identObj.Get(tokens.Cur().Str); obj != nil {
            if _,ok := obj.(*identObj.Struct); ok {
                return true
            }

            if _,ok := obj.(*identObj.Interface); ok {
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
func isNextType(tokens *token.Tokens) bool {
    tokens.SaveIdx()
    defer tokens.ResetIdx()
    return isNextType_(tokens)
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

    tokens.Next()
    generic := prsGeneric(tokens)
    isGeneric := generic.Str != ""

    s := identObj.DecStruct(name)

    if isGeneric {
        genericType := types.CreateGeneric(generic.Str)
        identObj.SetGeneric(&genericType)
        s.SetGeneric(&genericType)
    }

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

    tokens.Next()
    generic := prsGeneric(tokens)
    isGeneric := generic.Str != ""

    name := tokens.Cur()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    braceLPos := tokens.Next().Pos
    identObj.StartScope()
    I := identObj.DecInterface(name)
    identObj.CurSelfType = I.GetType()

    if isGeneric {
        genericType := types.CreateGeneric(generic.Str)
        identObj.SetGeneric(&genericType)
        // TODO: I.SetGeneric(&genericType)
    }

    heads := make([]ast.FnHead, 0)

    for tokens.Next().Type != token.BraceR {
        identObj.StartScope()
        fnHead := prsFnHead(tokens, true)
        identObj.EndScope()

        heads = append(heads, fnHead)
        I.AddFunc(fnHead.F)
    }

    braceRPos := tokens.Cur().Pos
    identObj.EndScope()
    identObj.CurSelfType = nil

    return &ast.DefInterface{ Pos: pos, Name: name, I: I, BraceLPos: braceLPos, BraceRPos: braceRPos, FnHeads: heads }
}

func prsEnum(tokens *token.Tokens) ast.Decl {
    pos := tokens.Cur().Pos

    name := tokens.Next()
    if name.Type != token.Name {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", name)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    tokens.Next()
    generic := prsGeneric(tokens)
    isGeneric := generic.Str != ""

    var idTyp types.Type = nil
    if tokens.Cur().Type != token.BraceL {
        idTyp = prsType(tokens)
        tokens.Next()
    } else {
        idTyp = types.CreateUint(types.Ptr_Size)
    }

    e := identObj.DecEnum(name)
    if isGeneric {
        genericType := types.CreateGeneric(generic.Str)
        identObj.SetGeneric(&genericType)
        e.SetGeneric(&genericType)
    }

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

    tokens.Next()
    generic := prsGeneric(tokens)
    isGeneric := generic.Str != ""

    DstName := tokens.Cur()
    if DstName.Type != token.Name && DstName.Type != token.Typename {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a name or typename but got %v\n", DstName)
        fmt.Fprintln(os.Stderr, "\t" + DstName.At())
        os.Exit(1)
    }

    var I *identObj.Interface = nil
    if tokens.Peek().Type == token.DefConst {
        tokens.Next()
        InterfaceName := tokens.Next()
        if InterfaceName.Type != token.Name {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a Name but got %v\n", InterfaceName)
            fmt.Fprintln(os.Stderr, "\t" + InterfaceName.At())
            os.Exit(1)
        }

        i,ok := identObj.Get(InterfaceName.Str).(*identObj.Interface)
        if !ok || i == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] interface \"%s\" is not defined\n", InterfaceName.Str)
            fmt.Fprintln(os.Stderr, "\t" + InterfaceName.At())
            os.Exit(1)
        }
        I = i
    }

    var dstType types.Type = nil
    var implementable identObj.Implementable = nil
    switch DstName.Type {
    case token.Name, token.Typename:
        obj := identObj.Get(DstName.Str)
        if obj == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not defined\n", DstName.Str)
            fmt.Fprintln(os.Stderr, "\t" + DstName.At())
            os.Exit(1)
        }
        implementable = obj.(identObj.Implementable)
        dstType = obj.GetType()

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] expected a Name or an TypeName but got %v\n", DstName)
        fmt.Fprintln(os.Stderr, "\t" + DstName.At())
        os.Exit(1)
    }

    identObj.CurSelfType = dstType

    braceLPos := tokens.Next().Pos
    if tokens.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected a \"{\" but got %v\n", tokens.Cur().Str)
        fmt.Fprintln(os.Stderr, "\t" + tokens.Cur().At())
        os.Exit(1)
    }
    identObj.StartScope()

    impl := identObj.CreateImpl(pos, I, dstType)
    implementable.AddImpl(impl)

    if isGeneric {
        genericType := types.CreateGeneric(generic.Str)
        identObj.SetGeneric(&genericType)
        // TODO: impl.SetGeneric(&genericType)
    }

    funcsReservedLen := 5
    if I != nil { funcsReservedLen = len(I.GetFuncs()) }
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
    identObj.EndScope()
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

    var f *identObj.Func = nil
    if isInterfaceFn && identObj.CurSelfType != nil {
        f = identObj.DecInterfaceFunc(name, isConst, identObj.CurSelfType)
    } else {
        f = identObj.DecFunc(name, isConst)
    }

    tokens.Next()
    generic := prsGeneric(tokens)
    isGeneric := generic.Str != ""
    if isGeneric {
        genericType := types.CreateGeneric(generic.Str)
        f.SetGeneric(&genericType)
        identObj.SetGeneric(&genericType)
    }

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

    return ast.FnHead{ Name: name, F: f, Args: argDecs, RetType: retType, IsConst: isConst, IsGeneric: isGeneric, Generic: generic }
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

    identObj.StartScope()
    block := prsBlock(tokens)
    identObj.EndScope()

    identObj.EndScope()

    def := ast.DefFn{ Pos: pos, FnHead: fnHead, Block: block }

    if fnHead.IsConst {
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

func prsOptionalSelfType(tokens *token.Tokens) types.Type {
    if tokens.Peek().Type == token.SelfType {
        tokens.Next()
        return createSelfType()
    }

    if tokens.Peek().Type == token.Mul && tokens.Peek2().Type == token.SelfType {
        tokens.Next()
        tokens.Next()
        return types.PtrType{ BaseType: createSelfType() } 
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
        return tokens.Next(), types.PtrType{ BaseType: createSelfType() } 
    }

    if tokens.Cur().Type == token.Self {
        name = tokens.Cur()
        typ = prsOptionalSelfType(tokens)
        if typ == nil {
            typ = createSelfType()
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

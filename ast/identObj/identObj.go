package identObj

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
    "gamma/cmpTime/constVal"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
)

type IdentObj interface {
    GetName() string
    GetType() types.Type
    GetPos() token.Pos
    Addr() addr.Addr
}

var curFunc *fn.Func = nil

func GetCurFunc() *fn.Func {
    return curFunc
}

func DecVar(name token.Token, t types.Type) vars.Var {
    checkName(name)

    if InGlobalScope() {
        v := vars.CreateGlobalVar(name, t)
        curScope.identObjs[name.Str] = &v
        return &v
    } else {
        v := vars.CreateLocal(name, t, curScope.frameSize, false, false)
        curScope.frameSize += v.GetType().Size()
        curScope.identObjs[name.Str] = &v
        return &v
    }
}

func DecArg(name token.Token, t types.Type) vars.Var {
    v := vars.CreateLocal(name, t, curScope.frameSize, true, false)
    curScope.frameSize += v.GetType().Size()
    curScope.identObjs[name.Str] = &v
    return &v
}

func DecArgFromStack(name token.Token, t types.Type, offset uint) vars.Var {
    v := vars.CreateLocal(name, t, offset, true, true)
    curScope.identObjs[name.Str] = &v
    return &v
}

func DecConst(name token.Token, t types.Type, val constVal.ConstVal) *consts.Const {
    checkName(name)

    c := consts.CreateConst(name, t, val)
    curScope.identObjs[name.Str] = &c
    return &c
}

func DecFunc(name token.Token) *fn.Func {
    checkName(name)

    f := fn.CreateFunc(name)
    curScope.identObjs[name.Str] = &f
    curFunc = &f
    return curFunc
}

func DecStruct(name token.Token, names []string, types []types.Type) *structDec.Struct {
    if InGlobalScope() {
        s := structDec.CreateStruct(name, names, types)
        curScope.identObjs[name.Str] = &s
        return &s
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare a struct in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
        return nil
    }
}

func AddBuildIn(name string, argtype types.Type, retType types.Type) {
    if !InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] AddBuildIn has to be called in the global scope")
        os.Exit(1)
    }

    f := fn.CreateFunc(token.Token{ Str: name })  
    f.SetRetType(retType)
    f.SetArgs([]types.Type{ argtype })
    curScope.identObjs[name] = &f
}

func SetRetType(retType types.Type) {
    if types.IsBigStruct(retType) {
        curScope.frameSize += types.Ptr_Size
    }
}

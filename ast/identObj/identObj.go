package identObj

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/ast/identObj/func"
    "gamma/ast/identObj/vars"
    "gamma/ast/identObj/consts"
    "gamma/ast/identObj/struct"
)

type IdentObj interface {
    GetName() string
    GetType() types.Type
    GetPos() token.Pos
    Addr(fieldNum int) string
}

func DecVar(name token.Token, t types.Type) vars.Var {
    if InGlobalScope() {
        v := vars.CreateGlobalVar(name, t)
        curScope.identObjs[name.Str] = &v
        return &v
    } else {
        v := vars.CreateLocal(name, t, curScope.frameSize)
        curScope.frameSize += v.GetType().Size()
        curScope.identObjs[name.Str] = &v
        return &v
    }
}

func DecConst(name token.Token, t types.Type) *consts.Const {
    checkName(name)

    c := consts.CreateConst(name, t)
    curScope.identObjs[name.Str] = &c
    return &c
}

func DecFunc(name token.Token) *fn.Func {
    checkName(name)

    f := fn.CreateFunc(name, nil)
    curScope.identObjs[name.Str] = &f
    return &f
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

func AddBuildIn(name string, argname string, argtype types.Type) {
    if !InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] AddBuildIn has to be called in the global scope")
        os.Exit(1)
    }

    f := fn.CreateFunc(
        token.Token{ Str: name, Type: token.Name },
        []types.Type{ argtype },
    )

    curScope.identObjs[name] = &f
}

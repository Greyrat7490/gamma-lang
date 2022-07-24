package identObj

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/identObj/func"
    "gorec/identObj/vars"
    "gorec/identObj/consts"
)

type IdentObj interface {
    Identobj()
}

func DecVar(name token.Token, t types.Type) vars.Var {
    if InGlobalScope() {
        v := vars.CreateGlobalVar(name, t)
        curScope.identObjs[name.Str] = &v
        return &v
    } else {
        v := vars.CreateLocal(name, t)
        curScope.identObjs[name.Str] = &v
        return &v
    }
}

func DecConst(name token.Token, t types.Type) *consts.Const {
    checkName(name)

    c := consts.Const{ Name: name, Type: t }
    curScope.identObjs[name.Str] = &c
    return &c
}

func DecFunc(name token.Token) *fn.Func {
    checkName(name)

    f := fn.CreateFunc(name)
    curScope.identObjs[name.Str] = &f
    return &f
}

func AddBuildIn(name string, argname string, argtype types.Type) {
    if !InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] AddBuildIn has to be called in the global scope")
        os.Exit(1)
    }

    f := fn.CreateFuncWithArgs(
        token.Token{ Str: name, Type: token.Name },
        []types.Type{ argtype },
    )

    curScope.identObjs[name] = &f
}

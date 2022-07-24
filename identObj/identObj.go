package identObj

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/identObj/func"
    "gorec/identObj/vars"
    "gorec/identObj/scope"
    "gorec/identObj/consts"
)

func DecVar(name token.Token, t types.Type) vars.Var {
    checkName(name)

    if scope.InGlobal() {
        v := vars.CreateGlobalVar(name, t)
        scope.AddVar(&v)
        return &v
    } else {
        v := vars.CreateLocal(name, t)
        scope.AddVar(&v)
        return &v
    }
}

func DecConst(name token.Token, t types.Type) *consts.Const {
    checkName(name)

    c := consts.Const{ Name: name, Type: t }
    scope.AddConst(&c)
    return &c
}

func DecFunc(name token.Token) *fn.Func {
    checkName(name)

    f := fn.CreateFunc(name)
    scope.AddFunc(&f)
    return &f
}


func checkName(name token.Token) {
    if name.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if scope.NameTaken(name.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] name \"%s\" is already taken in this scope\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
}

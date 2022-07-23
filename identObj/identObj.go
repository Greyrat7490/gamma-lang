package identObj

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
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


func checkName(name token.Token) {
    if name.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if scope.VarNameTaken(name.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is already declared in this scope\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if scope.ConstNameTaken(name.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] const \"%s\" is already declared in this scope\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
}

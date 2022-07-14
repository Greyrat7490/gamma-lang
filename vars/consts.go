package vars

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
)

type Const struct {
    Name token.Token
    Type types.Type
    Val token.Token
}


var globalConsts []Const

func GetConst(name string) *Const {
    for i := len(scopes)-1; i >= 0; i-- {
        for _, c := range scopes[i].consts {
            if c.Name.Str == name {
                return &c
            }
        }
    }

    for _, c := range globalConsts {
        if c.Name.Str == name {
            return &c
        }
    }

    return nil
}

func DefConst(name token.Token, conType types.Type, val token.Token) {
    if InGlobalScope() {
        if isGlobalVarDec(name.Str) {
            fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", name.Str)
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }

        if isGlobalConstDec(name.Str) {
            fmt.Fprintf(os.Stderr, "[ERROR] a const with the name \"%s\" is already declared\n", name.Str)
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }

        globalConsts = append(globalConsts, Const{ Name: name, Type: conType, Val: val })
    } else {
        if varInCurScope(name.Str) {
            fmt.Fprintf(os.Stderr, "[ERROR] local var \"%s\" is already declared in this scope\n", name.Str)
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }

        if constInCurScope(name.Str) {
            fmt.Fprintf(os.Stderr, "[ERROR] local const \"%s\" is already declared in this scope\n", name.Str)
            fmt.Fprintln(os.Stderr, "\t" + name.At())
            os.Exit(1)
        }

        scopes[len(scopes)-1].consts = append(scopes[len(scopes)-1].consts, Const{
            Name: name,
            Type: conType,
            Val: val,
        })
    }
}

func isGlobalConstDec(name string) bool {
    for _, c := range globalConsts {
        if c.Name.Str == name {
            return true
        }
    }

    return false
}

func constInCurScope(name string) bool {
    for _,c := range scopes[len(scopes)-1].consts {
        if c.Name.Str == name {
            return true
        }
    }

    return false
}

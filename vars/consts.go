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


var globalConsts []*Const

func GetConst(name string) *Const {
    scope := curScope

    for scope != nil {
        for i := len(scope.children)-1; i >= 0; i-- {
            for _, c := range scope.children[i].consts {
                if c.Name.Str == name {
                    return c
                }
            }
        }
        scope = scope.parent
    }

    for _, c := range globalConsts {
        if c.Name.Str == name {
            return c
        }
    }

    return nil
}

func DecConst(name token.Token, conType types.Type) *Const {
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

        c := Const{ Name: name, Type: conType }
        globalConsts = append(globalConsts, &c)
        return &c
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

        c := Const{ Name: name, Type: conType }
        curScope.consts = append(curScope.consts, &c)
        return &c
    }
}

func (c *Const) Define(val token.Token) {
    c.Val = val
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
    for _,c := range curScope.consts {
        if c.Name.Str == name {
            return true
        }
    }

    return false
}

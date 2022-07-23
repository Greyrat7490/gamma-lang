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


func GetConst(name string) *Const {
    scope := curScope

    for scope != nil {
        if c,ok := scope.consts[name]; ok {
            return c
        }

        scope = scope.parent
    }

    return nil
}

func DecConst(name token.Token, conType types.Type) *Const {
    if name.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if varNameTaken(name.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is already declared in this scope\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if constNameTaken(name.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] const \"%s\" is already declared in this scope\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    c := Const{ Name: name, Type: conType }
    curScope.consts[name.Str] = &c
    return &c
}

func (c *Const) Define(val token.Token) {
    c.Val = val
}

func constNameTaken(name string) bool {
    if _,ok := curScope.consts[name]; ok {
        return ok
    }

    return false
}

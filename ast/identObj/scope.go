package identObj

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
)

var curScope *Scope = &Scope{ identObjs: map[string]IdentObj{} }

type Scope struct {
    identObjs map[string]IdentObj
    parent *Scope
    innerSize uint
    frameSize uint
}

func GetCur() *Scope {
    return curScope
}

func GetFrameSize() uint {
    return curScope.frameSize + curScope.innerSize
}

func InGlobalScope() bool {
    return curScope.parent == nil
}

func StartScope() {
    curScope = &Scope{ parent: curScope, frameSize: curScope.frameSize, identObjs: map[string]IdentObj{} }
}

func EndScope() {
    if !InGlobalScope() {
        size := curScope.frameSize + curScope.innerSize
        if curScope.parent.innerSize < size {
            curScope.parent.innerSize = size
        }

        curScope = curScope.parent

        if InGlobalScope() {
            curScope.frameSize = 0
        }
    }
}

func Get(name string) IdentObj {
    scope := curScope

    for scope != nil {
        if f,ok := scope.identObjs[name]; ok {
            return f
        }

        scope = scope.parent
    }

    return nil
}

func GetGeneric(name string) *types.GenericType {
    if curFunc != nil {
        t := curFunc.GetGeneric()

        if t != nil && t.Name == name {
            return t
        }
    }

    return nil
}

func nameTaken(name string) bool {
    if _,ok := curScope.identObjs[name]; ok {
        return true
    }

    return false
}


func checkName(name token.Token) {
    if name.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if nameTaken(name.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] name \"%s\" is already taken in this scope\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
}

package identObj

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/loops"
    "gorec/conditions"
    "gorec/identObj/vars"
)

var curScope *Scope = &Scope{ identObjs: map[string]IdentObj{} }

type Scope struct {
    identObjs map[string]IdentObj
    parent *Scope
    innerSize int
}

func GetCur() *Scope {
    return curScope
}

func GetMaxFrameSize() int {
    return curScope.calcFrameSize()
}

func InGlobalScope() bool {
    return curScope.parent == nil
}

func StartScope() {
    curScope = &Scope{ parent: curScope, identObjs: map[string]IdentObj{} }
}

func EndScope() {
    if !InGlobalScope() {
        size := curScope.calcFrameSize()
        if curScope.parent.innerSize < size {
            curScope.parent.innerSize = size
        }

        curScope = curScope.parent

        if InGlobalScope() {
            vars.ResetLocalVarOffset()
            cond.ResetCount()
            loops.ResetCount()
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

func nameTaken(name string) bool {
    if _,ok := curScope.identObjs[name]; ok {
        return true
    }

    return false
}


func (s *Scope) calcFrameSize() (size int) {
    for _,v := range s.identObjs {
        if v,ok := v.(vars.Var); ok {
            size += v.GetType().Size()
        }
    }

    return size + curScope.innerSize
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

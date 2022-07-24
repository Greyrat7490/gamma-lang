package scope

import (
    "os"
    "fmt"
    "gorec/loops"
    "gorec/token"
    "gorec/types"
    "gorec/conditions"
    "gorec/identObj/func"
    "gorec/identObj/vars"
    "gorec/identObj/consts"
)

var curScope *Scope = &Scope{
    vars: map[string]vars.Var{},
    consts: map[string]*consts.Const{},
    funcs: map[string]*fn.Func{},
}

type Scope struct {
    vars map[string]vars.Var
    consts map[string]*consts.Const
    funcs map[string]*fn.Func
    parent *Scope
    innerSize int
}

func GetCur() *Scope {
    return curScope
}

func GetMaxFrameSize() int {
    return curScope.calcFrameSize()
}

func InGlobal() bool {
    return curScope.parent == nil
}

func Start() {
    curScope = &Scope{
        vars: map[string]vars.Var{},
        consts: map[string]*consts.Const{},
        funcs: map[string]*fn.Func{},
        parent: curScope,
    }
}

func End() {
    if !InGlobal() {
        size := curScope.calcFrameSize()
        if curScope.parent.innerSize < size {
            curScope.parent.innerSize = size
        }

        curScope = curScope.parent

        if InGlobal() {
            vars.ResetLocalVarOffset()
            cond.ResetCount()
            loops.ResetCount()
        }
    }
}


func AddVar(v vars.Var) {
    curScope.vars[v.GetName().Str] = v
}
func AddConst(c *consts.Const) {
    curScope.consts[c.Name.Str] = c
}
func AddFunc(f *fn.Func) {
    curScope.funcs[f.GetName().Str] = f
}
func AddBuildIn(name string, argname string, argtype types.Type) {
    if !InGlobal() {
        fmt.Fprintln(os.Stderr, "[ERROR] AddBuildIn has to be called in the global scope")
        os.Exit(1)
    }

    f := fn.CreateFuncWithArgs(
        token.Token{ Str: name, Type: token.Name },
        []types.Type{ argtype },
    )
    curScope.funcs[name] = &f
}


func GetVar(name string) vars.Var {
    scope := curScope

    for scope != nil {
        if v,ok := scope.vars[name]; ok {
            return v
        }

        scope = scope.parent
    }

    return nil
}

func GetConst(name string) *consts.Const {
    scope := curScope

    for scope != nil {
        if c,ok := scope.consts[name]; ok {
            return c
        }

        scope = scope.parent
    }

    return nil
}

func GetFunc(name string) *fn.Func {
    scope := curScope

    for scope != nil {
        if f,ok := scope.funcs[name]; ok {
            return f
        }

        scope = scope.parent
    }

    return nil
}

func NameTaken(name string) bool {
    if _,ok := curScope.vars[name]; ok {
        return true
    }
    if _,ok := curScope.consts[name]; ok {
        return true
    }
    if _,ok := curScope.funcs[name]; ok {
        return true
    }

    return false
}


func (s *Scope) calcFrameSize() (size int) {
    for _,v := range s.vars {
        size += v.GetType().Size()
    }

    return size + curScope.innerSize
}

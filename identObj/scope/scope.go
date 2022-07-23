package scope

import (
    "gorec/loops"
    "gorec/conditions"
    "gorec/identObj/vars"
    "gorec/identObj/consts"
)

var curScope *Scope = &Scope{
    vars: map[string]vars.Var{},
    consts: map[string]*consts.Const{},
}

type Scope struct {
    vars map[string]vars.Var
    consts map[string]*consts.Const
    parent *Scope
    innerSize int
}

func GetCur() *Scope {
    return curScope
}

func (s *Scope) GetMaxFrameSize() int {
    return s.calcFrameSize()
}

func InGlobal() bool {
    return curScope.parent == nil
}

func Start() {
    curScope = &Scope{
        vars: map[string]vars.Var{},
        consts: map[string]*consts.Const{},
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

func VarNameTaken(name string) bool {
    if _,ok := curScope.vars[name]; ok {
        return ok
    }

    return false
}

func ConstNameTaken(name string) bool {
    if _,ok := curScope.consts[name]; ok {
        return ok
    }

    return false
}


func (s *Scope) calcFrameSize() (size int) {
    for _,v := range s.vars {
        size += v.GetType().Size()
    }

    return size + curScope.innerSize
}

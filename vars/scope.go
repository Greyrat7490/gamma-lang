package vars

var curScope *Scope = &Scope{
    vars: map[string]Var{},
    consts: map[string]*Const{},
}

type Scope struct {
    vars map[string]Var
    consts map[string]*Const
    parent *Scope
    innerSize int
}

func GetCurScope() *Scope {
    return curScope
}

func (s *Scope) GetMaxFrameSize() int {
    return s.calcFrameSize()
}

func InGlobalScope() bool {
    return curScope.parent == nil
}

func StartScope() {
    curScope = &Scope{
        vars: map[string]Var{},
        consts: map[string]*Const{},
        parent: curScope,
    }
}

func EndScope() {
    if !InGlobalScope() {
        size := curScope.calcFrameSize()
        if curScope.parent.innerSize < size {
            curScope.parent.innerSize = size
        }

        curScope = curScope.parent

        if InGlobalScope() {
            localVarOffset = 0
        }
    }
}

func (s *Scope) calcFrameSize() (size int) {
    for _,v := range s.vars {
        size += v.GetType().Size()
    }

    return size + curScope.innerSize
}

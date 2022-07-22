package vars

var curScope *Scope = &Scope{
    vars: map[string]Var{},
    consts: map[string]*Const{},
}

type Scope struct {
    vars map[string]Var
    consts map[string]*Const
    parent *Scope
    children []Scope
    minFrameSize int
}

func GetCurScope() *Scope {
    return curScope
}

func (s *Scope) GetMaxFrameSize() int {
    maxInner := 0
    for _,c := range s.children {
        size := c.GetMaxFrameSize()
        if size > maxInner {
            maxInner = size
        }
    }

    return s.minFrameSize + maxInner
}

func InGlobalScope() bool {
    return curScope.parent == nil
}

func CreateScope() {
    curScope.children = append(curScope.children, Scope{
        vars: map[string]Var{},
        consts: map[string]*Const{},
        parent: curScope,
    })
    curScope = &curScope.children[len(curScope.children)-1]
}

func EndScope() {
    if !InGlobalScope() {
        curScope = curScope.parent

        if InGlobalScope() {
            localVarOffset = 0
        }
    }
}

package vars

var scope Scope = Scope{ vars: make(map[string]Var), consts: map[string]*Const{} }
var curScope *Scope = &scope

type Scope struct {
    vars map[string]Var
    consts map[string]*Const
    parent *Scope
    children []Scope
    maxSize int
}

func InGlobalScope() bool {
    return curScope.parent == nil
}

func CreateScope() {
    curScope.children = append(curScope.children, Scope{
        vars: make(map[string]Var),
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

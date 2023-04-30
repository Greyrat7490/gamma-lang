package identObj

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime/constVal"
    "gamma/ast/identObj/vars"
)

type Scope struct {
    identObjs map[string]IdentObj
    parent *Scope
    children []Scope
}

var globalScope = Scope{ identObjs: make(map[string]IdentObj), children: make([]Scope, 0) }
var curScope = &globalScope
var stackSize uint = 0

func (s *Scope) ArgsSize() uint {
    size := uint(0)

    for _,obj := range s.identObjs {
        if !types.IsBigStruct(obj.GetType()) {
            size += obj.GetType().Size()
        }
    }

    return size
}

func (s *Scope) getInnerSize(size uint) uint {
    for _,obj := range s.identObjs {
        if _,ok := obj.(*vars.LocalVar); ok {
            size += obj.GetType().Size()
        }
    }

    for _,s := range s.children {
        size += s.getInnerSize(size)
    }

    return size
}

func (s *Scope) GetInnerSize() uint {
    if len(s.children) != 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (internal) expected scope of function to have 1 child scope, but got %d\n", len(s.children))
        os.Exit(1)
    }

    size := s.children[0].getInnerSize(0)

    // framesize has to be the multiple of 16byte
    return (size + 15) & ^uint(15)
}

func InGlobalScope() bool {
    return curScope.parent == nil
}

func StartScope() {
    scope := Scope{ parent: curScope, identObjs: make(map[string]IdentObj), children: make([]Scope, 0) }
    curScope.children = append(curScope.children, scope)
    curScope = &curScope.children[len(curScope.children)-1]
}

func EndScope() {
    if !InGlobalScope() {
        curScope = curScope.parent
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

func GetStackSize() uint {
    return stackSize
}

func IncStackSize(t types.Type) {
    stackSize += t.Size()
}

func ResetStackSize() {
    stackSize = 0
}

func (scope *Scope) nameTaken(name string) bool {
    if _,ok := scope.identObjs[name]; ok {
        return true
    }

    return false
}


func (scope *Scope) checkName(name token.Token) {
    if name.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    if scope.nameTaken(name.Str) {
        fmt.Fprintf(os.Stderr, "[ERROR] name \"%s\" is already taken in this scope\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }
}

func AddBuildIn(name string, argtype types.Type, retType types.Type) {
    f := CreateFunc(token.Token{ Str: name })  
    f.SetRetType(retType)
    f.SetArgs([]types.Type{ argtype })
    globalScope.identObjs[name] = &f
}

func AddGenBuildIn(name string, genericName string, argtype types.Type, retType types.Type) {
    f := CreateFunc(token.Token{ Str: name })  
    f.SetGeneric(&types.GenericType{Name: genericName, UsedTypes: make([]types.Type, 0)})
    f.SetRetType(retType)
    if argtype != nil {
        f.SetArgs([]types.Type{ argtype })
    }
    globalScope.identObjs[name] = &f
}

func DecVar(name token.Token, t types.Type) vars.Var {
    curScope.checkName(name)

    if InGlobalScope() {
        v := vars.CreateGlobalVar(name, t)
        curScope.identObjs[name.Str] = &v
        return &v
    } else {
        v := vars.CreateLocal(name, t)
        curScope.identObjs[name.Str] = &v
        return &v
    }
}

func DecConst(name token.Token, t types.Type, val constVal.ConstVal) *Const {
    curScope.checkName(name)

    c := CreateConst(name, t, val)
    curScope.identObjs[name.Str] = &c
    return &c
}

func DecFunc(name token.Token) *Func {
    curScope.checkName(name)

    f := CreateFunc(name)
    StartScope()
    f.Scope = curScope

    curScope.parent.identObjs[name.Str] = &f
    curFunc = &f

    return curFunc
}

func DecStruct(name token.Token, names []string, types []types.Type) *Struct {
    if InGlobalScope() {
        s := CreateStruct(name, names, types)
        curScope.identObjs[name.Str] = &s
        return &s
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare a struct in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
        return nil
    }
}

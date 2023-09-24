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
    unnamedVars uint
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

func (s *Scope) getInnerSize() uint {
    size := uint(0)
    for _,obj := range s.identObjs {
        if _,ok := obj.(*vars.LocalVar); ok {
            size += obj.GetType().Size()
        }
    }

    for _,s := range s.children {
        size += s.getInnerSize()
    }

    return size
}

func (s *Scope) GetInnerSize() uint {
    if len(s.children) != 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (internal) expected scope of function to have 1 child scope, but got %d\n", len(s.children))
        os.Exit(1)
    }

    size := s.children[0].getInnerSize()

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
    f := CreateFunc(token.Token{ Str: name }, false)  
    f.SetRetType(retType)
    f.SetArgs([]types.Type{ argtype })
    globalScope.identObjs[name] = &f
}

func AddGenBuildIn(name string, genericName string, argtype types.Type, retType types.Type) {
    f := CreateFunc(token.Token{ Str: name }, false)
    f.SetGeneric(&types.GenericType{Name: genericName, UsedTypes: make([]types.Type, 0)})
    f.SetRetType(retType)
    if argtype != nil {
        f.SetArgs([]types.Type{ argtype })
    }
    globalScope.identObjs[name] = &f
}

func AddPrimitives() {
    AddPrimitive(types.CreateUint(types.U64_Size))
    AddPrimitive(types.CreateUint(types.U32_Size))
    AddPrimitive(types.CreateUint(types.U16_Size))
    AddPrimitive(types.CreateUint(types.U8_Size))

    AddPrimitive(types.CreateInt(types.I64_Size))
    AddPrimitive(types.CreateInt(types.I32_Size))
    AddPrimitive(types.CreateInt(types.I16_Size))
    AddPrimitive(types.CreateInt(types.I8_Size))

    AddPrimitive(types.BoolType{ Interfaces: make(map[string]types.InterfaceType) })
    AddPrimitive(types.CharType{ Interfaces: make(map[string]types.InterfaceType) })
    AddPrimitive(types.StrType{ Interfaces: make(map[string]types.InterfaceType) })
}                                

func DecVar(name token.Token, t types.Type) vars.Var {
    if name.Type == token.UndScr {
        return ReserveSpace(t)
    }

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

func DecFunc(name token.Token, isConst bool) *Func {
    curScope.parent.checkName(name)

    f := CreateFunc(name, isConst)
    f.Scope = curScope

    curScope.parent.identObjs[name.Str] = &f
    curFunc = &f

    return curFunc
}

func DecInterfaceFunc(name token.Token, isConst bool, receiver types.Type) *Func {
    curScope.parent.checkName(name)

    f := CreateInterfaceFunc(name, isConst, receiver)
    f.Scope = curScope

    curScope.parent.identObjs[name.Str] = &f
    curFunc = &f

    return curFunc
}

func DecInterface(name token.Token) *Interface {
    curScope.parent.checkName(name)

    I := CreateInterface(name)
    I.scope = curScope

    curScope.parent.identObjs[name.Str] = &I

    return &I
}

func DecStruct(name token.Token) *Struct {
    if !InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare a struct in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
        return nil
    }

    curScope.checkName(name)

    s := CreateStruct(name)
    curScope.identObjs[name.Str] = &s
    return &s
}

func DecEnum(name token.Token, idType types.Type, names []string, types []types.Type) *Enum {
    if !InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare an enum in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
        return nil
    }

    curScope.checkName(name)

    e := CreateEnum(name, idType, names, types)
    curScope.identObjs[name.Str] = &e
    return &e
}

func ReserveSpace(t types.Type) vars.Var {
    if types.IsBigStruct(t) {
        name := fmt.Sprintf("_reserved%d", curScope.unnamedVars)
        curScope.unnamedVars += 1

        v := vars.CreateLocal(token.Token{ Str: name }, t)
        curScope.identObjs[name] = &v
        return &v
    } else {
        name := "_"
        v := vars.CreateLocal(token.Token{ Str: name }, t)

        if obj,ok := curScope.identObjs[name]; !ok || obj.GetType().Size() < t.Size() {
            curScope.identObjs[name] = &v
        }

        return &v
    }
}

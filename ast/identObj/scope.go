package identObj

import (
	"fmt"
	"gamma/ast/identObj/vars"
	"gamma/cmpTime/constVal"
	"gamma/token"
	"gamma/types"
	"gamma/types/addr"
	"os"
)

type Scope struct {
    identObjs map[string]IdentObj
    implObj map[string]*Implementable
    parent *Scope
    children []Scope
    unnamedVars uint
    lastReserved *ReservedSpace
    reservedSpace uint
}

type ReservedSpace struct {
    used bool 
    address addr.Addr
    typ types.Type
}

var globalScope = Scope{ identObjs: make(map[string]IdentObj), implObj: make(map[string]*Implementable), children: make([]Scope, 0) }
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

    size += s.reservedSpace

    return size
}

func (s *Scope) GetInnerSize() uint {
    if len(s.children) != 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (internal) expected scope of function to have 1 child scope, but got %d\n", len(s.children))
        os.Exit(1)
    }

    size := s.children[0].getInnerSize()

    // framesize has to be the multiple of 16bits
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

func CreateImplObj(t types.Type) *Implementable {
    if obj,ok := globalScope.implObj[t.String()]; ok {
        return obj
    }

    obj := &Implementable{ dstType: t, impls: make([]Impl, 0, 1), interfaces: make([]string, 0, 1) }
    globalScope.implObj[t.String()] = obj
    return obj
}

func GetImplObj(name string) *Implementable {
    if obj,ok := globalScope.implObj[name]; ok {
        return obj
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
    f := CreateFunc(token.Token{ Str: name }, false, nil)  
    f.SetRetType(retType)
    f.SetArgs([]types.Type{ argtype })
    globalScope.identObjs[name] = &f
}

func AddGenBuildIn(name string, genericName string, argtype types.Type, retType types.Type) {
    f := CreateFunc(token.Token{ Str: name }, false, nil)
    f.SetGeneric(&types.GenericType{Name: genericName, UsedInsetTypes: make([]types.Type, 0)})
    f.SetRetType(retType)
    if argtype != nil {
        f.SetArgs([]types.Type{ argtype })
    }
    globalScope.identObjs[name] = &f
}


func DecVar(name token.Token, t types.Type) vars.Var {
    if name.Type == token.UndScr && !InGlobalScope() {
        return createUnnamedVar(t)
    }

    curScope.checkName(name)

    if InGlobalScope() {
        v := vars.CreateGlobalVar(name, t)
        curScope.identObjs[name.Str] = &v
        return &v
    } else {
        v := vars.CreateLocal(name, t)
        curScope.identObjs[name.Str] = &v
        reuseSpace(&v)
        return &v
    }
}

func DecConst(name token.Token, t types.Type, val constVal.ConstVal) *Const {
    curScope.checkName(name)

    c := CreateConst(name, t, val)
    curScope.identObjs[name.Str] = &c
    return &c
}

func DecFunc(name token.Token, isConst bool, fnSrc types.Type) *Func {
    curScope.parent.checkName(name)

    f := CreateFunc(name, isConst, fnSrc)
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

func DecEnum(name token.Token) *Enum {
    if !InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] you can only declare an enum in the global scope")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
        return nil
    }

    curScope.checkName(name)

    e := CreateEnum(name)
    curScope.identObjs[name.Str] = &e
    return &e
}

func ReserveSpace(t types.Type) *addr.Addr {
    if types.IsBigStruct(t) {
        if curScope.lastReserved != nil && !curScope.lastReserved.used {
            if curScope.lastReserved.typ.Size() < t.Size() {
                curScope.lastReserved = &ReservedSpace{ used: false, typ: t }
                curScope.reservedSpace += t.Size() - curScope.lastReserved.typ.Size()
            }

            return &curScope.lastReserved.address
        }

        curScope.lastReserved = &ReservedSpace{ used: false, typ: t }
        curScope.reservedSpace += t.Size()
        return &curScope.lastReserved.address
    }

    return nil
}

func AllocReservedSpaceIfNeeded(t types.Type, reservedSpace *addr.Addr) {
    if reservedSpace.BaseAddr == "" {
        IncStackSize(t)
        *reservedSpace = addr.Addr{ BaseAddr: "rbp", Offset: -int64(stackSize) }
    }
}

func reuseSpace(v vars.Var) {
    if curScope.lastReserved == nil {
        return
    }

    if types.Equal(v.GetType(), curScope.lastReserved.typ) {
        curScope.reservedSpace -= v.GetType().Size()
        curScope.lastReserved.used = true
        curScope.lastReserved.address = v.Addr()
    }
}

func createUnnamedVar(t types.Type) vars.Var {
    if types.IsBigStruct(t) {
        name := fmt.Sprintf("_%d", curScope.unnamedVars)
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

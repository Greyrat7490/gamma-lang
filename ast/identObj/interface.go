package identObj

import (
    "os"
    "fmt"
    "reflect"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type Interface struct {
    decPos token.Pos
    name string
    scope *Scope
    typ types.InterfaceType
    funcs []Func
}

var CurSelfType types.Type = nil

func CreateInterface(name token.Token) Interface {
    return Interface{ decPos: name.Pos, name: name.Str, typ: types.InterfaceType{ Name: name.Str }, funcs: make([]Func, 0) }
}

func (i *Interface) GetName() string {
    return i.name
}

func (i *Interface) GetType() types.Type {
    return i.typ
}

func (i *Interface) GetPos() token.Pos {
    return i.decPos
}

func (i *Interface) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) Cannot get the addr of an interface definition")
    os.Exit(1)
    return addr.Addr{}
}

func (i *Interface) GetFuncs() []types.FuncType {
    return i.typ.Funcs
}

func (i *Interface) GetFunc(name string) *Func {
    for _,f := range i.funcs {
        if f.name == name {
            return &f
        }       
    }

    return nil
}

func (i *Interface) AddFunc(f *Func) {
    i.typ.Funcs = append(i.typ.Funcs, f.typ)
    i.funcs = append(i.funcs, *f)
}

func (i *Interface) GetVTableOffset(funcName string) uint {
    offset := uint(0)
    for _,f := range i.funcs {
        if f.name == funcName {
            return offset
        }

        if len(f.typ.Args) > 0 && isSelfType(f.typ.Args[0], i.typ) {
            offset += 8
        }
    }

    return 0
}

func (i *Interface) AddImpl(impl Impl) { 
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) Cannot add impl to an interface")
    os.Exit(1)
}

func (i *Interface) HasInterface(name string) bool {
    return name == i.name
}

func (i *Interface) GetFuncNames() []string {
    res := make([]string, 0, len(i.funcs)) 

    for _,f := range i.funcs {
        res = append(res, f.GetName())
    }

    return res
}


type Impl struct {
    decPos token.Pos
    interface_ *Interface // can be nil
    dstType types.Type
    scope *Scope
}

func CreateImpl(decPos token.Pos, interface_ *Interface, dstType types.Type) Impl {
    return Impl{ decPos: decPos, interface_: interface_, dstType: dstType, scope: curScope }
}

func (i *Impl) HasInterface() bool {
    return i.interface_ != nil
}

func (i *Impl) GetInterfaceName() string {
    if i.interface_ == nil {
        return ""
    }

    return i.interface_.name
}

func (i *Impl) GetImplName() string {
    return i.dstType.String()
}

func (i *Impl) GetInterfaceFuncs() []types.FuncType {
    if i.interface_ == nil {
        return nil
    }

    return i.interface_.typ.Funcs
}

func (i *Impl) GetVTableFuncNames() []string {
    if i.interface_ == nil {
        return nil
    }

    names := make([]string, 0, len(i.interface_.typ.Funcs))

    for _,f := range i.interface_.typ.Funcs {
        if len(f.Args) > 0 && isSelfType(f.Args[0], i.interface_.typ) {
            names = append(names, f.Name)
        }
    }

    return names
}

func (i *Impl) GetInterfaceFuncNames() []string {
    if i.interface_ == nil {
        return nil
    }

    names := make([]string, 0, len(i.interface_.typ.Funcs))

    for _,f := range i.interface_.typ.Funcs {
        names = append(names, f.Name)
    }

    return names
}

func (i *Impl) GetInterfaceFuncPos(name string) token.Pos {
    if i.interface_ != nil {
        for idx, f := range i.interface_.typ.Funcs {
            if f.Name == name {
                return i.interface_.funcs[idx].decPos
            }
        }
    }

    return token.Pos{}
}

func (i *Impl) GetInterfacePos() token.Pos {
    if i.interface_ == nil {
        return token.Pos{}
    }

    return i.interface_.decPos
}

func (i *Impl) GetFunc(name string) *Func {
    if obj,ok := i.scope.identObjs[name]; ok {
        if f,ok := obj.(*Func); ok {
            return f
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] (internal) the scope of Impl should only contain funcs but got %v\n", reflect.TypeOf(f))
            os.Exit(1)
        }
    }

    return nil
}

func isSelfType(t types.Type, interfaceType types.InterfaceType) bool {
    switch t := t.(type) {
    case types.InterfaceType:
        return types.Equal(t, interfaceType)

    case types.PtrType:
        return isSelfType(t.BaseType, interfaceType)

    default:
        return false
    }
}

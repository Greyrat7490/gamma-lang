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
    generic *Generic
    funcs []Func
}

var CurSelfType types.Type = nil

func CreateInterface(name token.Token, generic *Generic) Interface {
    interfaceType := types.CreateInterfaceType(name.Str)
    if generic != nil {
        interfaceType.Generic = generic.Typ
    }
    return Interface{ decPos: name.Pos, name: name.Str, typ: interfaceType, funcs: make([]Func, 0), generic: generic }
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

func (i *Interface) GetFunc(name string) *Func {
    for _,f := range i.funcs {
        if f.name == name {
            return &f
        }       
    }

    return nil
}

func (i *Interface) GetGeneric() *Generic {
    return i.generic
}

func (i *Interface) IsGeneric() bool {
    return i.generic != nil
}

func (i *Interface) AddTypeToGeneric(typ types.Type) {
   AddTypeToGeneric(i.generic, typ) 
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

        if len(f.typ.Args) > 0 && types.IsSelfType(f.typ.Args[0], i.typ) {
            offset += 8
        }
    }

    return 0
}

func (i *Interface) GetFuncNames() []string {
    res := make([]string, 0, len(i.funcs)) 

    for _,f := range i.funcs {
        res = append(res, f.GetName())
    }

    return res
}

func (i *Interface) SetInsetType(insetType types.Type) {
    types.SetCurInsetType(i.generic.Typ, insetType)
}


type Impl struct {
    decPos token.Pos
    interface_ *Interface               // can be nil
    interfaceType *types.InterfaceType  // can be nil
    dstType types.Type
    generic *Generic                    // can be nil
    scope *Scope
}

func CreateImpl(decPos token.Pos, interfaceType *types.InterfaceType, dstType types.Type, generic *Generic) Impl {
    if interfaceType != nil {
        interface_,ok := Get(interfaceType.Name).(*Interface)
        if !ok {
            fmt.Fprintf(os.Stderr, "[ERROR] (internal) interface %v is not defined correct\n", interfaceType)
            fmt.Fprintln(os.Stderr, "\t" + decPos.At())
            os.Exit(1)
        }
        return Impl{ decPos: decPos, interface_: interface_, interfaceType: interfaceType, dstType: dstType, scope: curScope, generic: generic }
    }

    return Impl{ decPos: decPos, interfaceType: interfaceType, dstType: dstType, scope: curScope, generic: generic }
}

func createImplFromGeneric(src types.Type, impl *Impl) Impl {
    impl.AddTypeToGeneric(src)

    newScope := *impl.scope
    newScope.identObjs = make(map[string]IdentObj, len(newScope.identObjs))
    newScope.implObj = make(Implementations, len(newScope.implObj))

    newImpl := *impl
    newImpl.scope = &newScope

    for _,obj := range impl.scope.identObjs {
        if f,ok := obj.(*Func); ok {
            newScope.identObjs[f.name] = f.replaceGeneric(src)
        }
    }

    return newImpl
}

func (i *Impl) HasInterface() bool {
    return i.interfaceType != nil
}

func (i *Impl) GetInterfaceName() string {
    if i.interfaceType == nil {
        return ""
    }

    return i.interfaceType.Name
}

func (i *Impl) GetInterfaceType() types.InterfaceType {
    if i.interfaceType == nil {
        return types.InterfaceType{}
    }

    return *i.interfaceType
}

func (i *Impl) GetDstType() types.Type {
    return i.dstType
}

func (i *Impl) GetGeneric() *Generic {
    return i.generic
}

func (i *Impl) IsGeneric() bool {
    return i.generic != nil
}

func (i *Impl) AddTypeToGeneric(typ types.Type) {
   AddTypeToGeneric(i.generic, typ) 
}

func (i *Impl) SetInsetType(insetType types.Type) {
    types.SetCurInsetType(i.generic.Typ, insetType)
}

func (i *Impl) GetInterfaceFuncs() []types.FuncType {
    if i.interfaceType == nil {
        return nil
    }

    return i.interfaceType.Funcs
}

func (i *Impl) GetVTableFuncNames() []string {
    if i.interfaceType == nil {
        return nil
    }

    names := make([]string, 0, len(i.interfaceType.Funcs))

    
    for _,f := range i.interfaceType.Funcs {
        if len(f.Args) > 0 && types.IsSelfType(f.Args[0], *i.interfaceType) {
            names = append(names, f.Name)
        }
    }

    return names
}

func (i *Impl) GetInterfaceFuncNames() []string {
    if i.interfaceType == nil {
        return nil
    }

    names := make([]string, 0, len(i.interfaceType.Funcs))

    for _,f := range i.interfaceType.Funcs {
        names = append(names, f.Name)
    }

    return names
}

func (i *Impl) GetInterfaceFuncPos(name string) token.Pos {
    if i.interfaceType != nil {
        for _,f := range i.interface_.funcs {
            if f.name == name {
                return f.decPos
            }
        }
    }

    return token.Pos{}
}

func (i *Impl) GetInterfacePos() token.Pos {
    if i.interfaceType != nil {
        return i.interface_.GetPos()
    }

    return i.decPos
}

func (i *Impl) GetFunc(name string) *Func {
    if obj,ok := i.scope.identObjs[name]; ok {
        if f,ok := obj.(*Func); ok {
            return f
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] (internal) the scope of Impl should only contain funcs but got %v\n", reflect.TypeOf(f))
            fmt.Fprintln(os.Stderr, "\t" + i.decPos.At())
            os.Exit(1)
        }
    }

    return nil
}

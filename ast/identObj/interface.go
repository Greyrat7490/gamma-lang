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
    funcPos []token.Pos
}

var CurImplStruct *types.StructType = nil

func CreateInterface(name token.Token) Interface {
    return Interface{ decPos: name.Pos, name: name.Str, typ: types.InterfaceType{ Name: name.Str }, funcPos: make([]token.Pos, 0) }
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

func (i *Interface) AddFunc(f Func) {
    i.typ.Funcs = append(i.typ.Funcs, f.typ)
    i.funcPos = append(i.funcPos, f.decPos)
}

type Impl struct {
    decPos token.Pos
    interface_ *Interface
    struct_ *Struct
    scope *Scope
}

func CreateImpl(decPos token.Pos, interface_ *Interface, struct_ *Struct) Impl {
    return Impl{ decPos: decPos, interface_: interface_, struct_: struct_, scope: curScope }
}

func (i *Impl) GetInterfaceName() string {
    return i.interface_.name
}

func (i *Impl) GetStructName() string {
    return i.struct_.name
}

func (i *Impl) GetInterfaceFuncs() []types.FuncType {
    return i.interface_.typ.Funcs
}

func (i *Impl) GetInterfaceFuncNames() []string {
    names := make([]string, 0, len(i.interface_.typ.Funcs))

    for _,f := range i.interface_.typ.Funcs {
        names = append(names, f.Name)
    }

    return names
}

func (i *Impl) GetInterfaceFuncPos(name string) token.Pos {
    for idx, f := range i.interface_.typ.Funcs {
        if f.Name == name {
            return i.interface_.funcPos[idx]
        }
    }

    return token.Pos{}
}

func (i *Impl) GetInterfacePos() token.Pos {
    return i.interface_.decPos
}

func (i *Impl) GetMethod(name string) *Func {
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

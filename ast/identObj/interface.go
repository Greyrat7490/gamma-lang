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
    Funcs []Func
}

var CurImplStruct *types.StructType = nil

func CreateInterface(name token.Token) Interface {
    return Interface{ decPos: name.Pos, name: name.Str, Funcs: make([]Func, 0) }
}

func (i *Interface) GetName() string {
    return i.name
}

func (i *Interface) GetType() types.Type {
    return nil
}

func (i *Interface) GetPos() token.Pos {
    return i.decPos
}

func (i *Interface) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) Cannot get the addr of an interface definition")
    os.Exit(1)
    return addr.Addr{}
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

func (i *Impl) GetInterfaceFuncs() []Func {
    return i.interface_.Funcs
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

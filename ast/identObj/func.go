package identObj

import (
	"gamma/token"
	"gamma/types"
	"gamma/types/addr"
)

type Func struct {
    decPos token.Pos
    name string
    typ types.FuncType
    retAddr addr.Addr   // TODO remove
    Scope *Scope
    receiver types.Type
    isConst bool
}

var curFunc *Func = nil

func GetCurFunc() *Func {
    return curFunc
}

func CreateFunc(name token.Token, isConst bool) Func {
    return Func{ name: name.Str, decPos: name.Pos, isConst: isConst, typ: types.FuncType{ Name: name.Str } }
}

func CreateInterfaceFunc(name token.Token, isConst bool, receiver types.Type) Func {
    return Func{ name: name.Str, decPos: name.Pos, isConst: isConst, receiver: receiver, typ: types.FuncType{ Name: name.Str } }
}

func (f *Func) GetArgs() []types.Type {
    return f.typ.Args
}

func (f *Func) GetName() string {
    return f.name
}

func (f *Func) GetType() types.Type {
    return f.typ
}

func (f *Func) GetGeneric() *types.GenericType {
    return f.typ.Generic
}

func (f *Func) GetRetType() types.Type {
    return f.typ.Ret
}

func (f *Func) GetPos() token.Pos {
    return f.decPos
}

func (f *Func) Addr() addr.Addr {
    return addr.Addr{ BaseAddr: f.GetMangledName() }
}

func (f *Func) GetRetAddr() addr.Addr {
    return f.retAddr
}

func (f *Func) GetMangledName() string {
    name := f.name

    if f.receiver != nil {
        name = f.receiver.String() + "." + name
    }

    if f.GetGeneric() != nil {
        name += "$" + f.typ.Generic.CurUsedType.String()
    }

    return name
}


func (f *Func) SetRetType(typ types.Type) {
    f.typ.Ret = typ
}

func (f *Func) SetRetAddr(addr addr.Addr) {
    f.retAddr = addr
}

func (f *Func) SetArgs(args []types.Type) {
    f.typ.Args = args
}

func (f *Func) SetGeneric(generic *types.GenericType) {
    f.typ.Generic = generic
}

func (f *Func) UpdateReceiver(recv types.Type) *Func {
    res := *f
    res.typ.Args = make([]types.Type, len(res.typ.Args))
    copy(res.typ.Args, f.typ.Args)
    res.typ.Args[0] = recv
    res.receiver = recv
    return &res
}

func (f *Func) IsGeneric() bool {
    return f.typ.Generic != nil
}

func (f *Func) AddTypeToGeneric(typ types.Type) {
    for _,t := range f.typ.Generic.UsedTypes {
        if types.Equal(typ, t) { return }
    }

    f.typ.Generic.UsedTypes = append(f.typ.Generic.UsedTypes, typ)
}

func (f Func) String() string {
    return f.typ.String()
}

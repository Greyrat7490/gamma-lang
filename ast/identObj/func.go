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
    Generic *Generic
    fnSrc types.Type
    hasSrcObj bool
    isConst bool
}

var curFunc *Func = nil

func GetCurFunc() *Func {
    return curFunc
}

func CreateFunc(name token.Token, isConst bool, fnSrc types.Type, generic *Generic) Func {
    if generic != nil {
        return Func{ name: name.Str, decPos: name.Pos, isConst: isConst, fnSrc: fnSrc, typ: types.FuncType{ Name: name.Str, Generic: generic.Typ }, Generic: generic }
    }

    return Func{ name: name.Str, decPos: name.Pos, isConst: isConst, fnSrc: fnSrc, typ: types.FuncType{ Name: name.Str } }
}

func CreateUnresolvedFunc(name string) Func {
    return Func{ typ: types.FuncType{ Ret: types.CreateInferType(nil) } }
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

func (f *Func) GetGeneric() types.GenericType {
    return f.typ.Generic
}

func (f *Func) GetUsedInsetTypes() []types.Type {
    return f.Generic.UsedInsetTypes
}

func (f *Func) RmDuplInsetTypes() {
    f.Generic.RemoveDuplTypes()
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

    if f.fnSrc != nil {
        name = f.fnSrc.GetMangledName() + "." + name
    }

    if f.IsGeneric() {
        name += "$" + f.typ.Generic.GetMangledName()
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

    if f.fnSrc != nil && len(args) > 0 {
        if t,ok := args[0].(types.PtrType); ok {
            f.hasSrcObj = types.Equal(f.fnSrc, t.BaseType)
        } else {
            f.hasSrcObj = types.Equal(f.fnSrc, args[0])
        }
    }
}

func (f *Func) ResolveFnSrc(t types.Type) {
    if f.fnSrc != nil && types.IsResolvable(f.typ.Args[0]) {
        f.typ.Args[0] = t
        f.fnSrc = t
    }
}

func (f *Func) IsGeneric() bool {
    return f.Generic != nil
}

func (f *Func) IsUnresolved() bool {
    return f.name == ""
}

func (f *Func) GetSrcObj() types.Type {
    if f.hasSrcObj {
        return f.typ.Args[0]
    }
    return nil
}

func (f Func) ResolveInferedTypes(typ types.Type) *Func {
    if types.IsResolvable(f.fnSrc) {
        f.fnSrc = typ
    }

    for i := range f.typ.Args {
        if types.IsResolvable(f.typ.Args[i]) {
            f.typ.Args[i] = typ
        }
    }

    if types.IsResolvable(f.typ.Ret) {
        f.typ.Ret = typ
    }

    if types.IsResolvable(f.typ.Generic.SetType) {
        f.typ.Generic.SetType = typ
    }

    AddTypeToGeneric(f.Generic, typ)

    return &f
}

func (f Func) replaceGeneric(typ types.Type) *Func {
    f.fnSrc = types.ReplaceGeneric(f.fnSrc, typ)

    args := f.typ.Args
    f.typ.Args = make([]types.Type, len(args))
    copy(f.typ.Args, args)

    for i,a := range f.typ.Args {
        f.typ.Args[i] = types.ReplaceGeneric(a, typ)
    }

    f.typ.Ret = types.ReplaceGeneric(f.typ.Ret, typ)
    f.typ.Generic.SetType = typ

    return &f
}

func (f *Func) ResolveGeneric(typ types.Type) *Func {
    if !f.IsGeneric() { return f }

    f = f.replaceGeneric(typ)

    AddTypeToGeneric(f.Generic, typ)
    return f
}

func (f *Func) SetInsetType(insetType types.Type) {
    types.SetCurInsetType(f.Generic.Typ, insetType)
}

func (f Func) String() string {
    return f.typ.String()
}

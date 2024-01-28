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
    fnSrc types.Type
    hasSrcObj bool
    isConst bool
}

var curFunc *Func = nil

func GetCurFunc() *Func {
    return curFunc
}

func CreateFunc(name token.Token, isConst bool, fnSrc types.Type) Func {
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

    if f.fnSrc != nil {
        name = f.fnSrc.GetMangledName() + "." + name
    }

    if f.GetGeneric() != nil {
        name += "$" + f.typ.Generic.CurInsetType.GetMangledName()
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

func (f *Func) SetGeneric(generic *types.GenericType) {
    f.typ.Generic = generic
}

func (f *Func) ResolveFnSrc(t types.Type) {
    if f.fnSrc != nil && types.IsResolvable(f.typ.Args[0]) {
        f.typ.Args[0] = t
        f.fnSrc = t
    }
}

func (f *Func) IsGeneric() bool {
    return f.typ.Generic != nil
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

func (f *Func) AddTypeToGeneric(typ types.Type) {
   AddTypeToGeneric(f.typ.Generic, typ) 
}

func (f Func) String() string {
    return f.typ.String()
}

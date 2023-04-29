package identObj

import (
	"os"
	"fmt"
	"gamma/token"
	"gamma/types"
	"gamma/types/addr"
)

type Func struct {
    decPos token.Pos
    name string
    generic *types.GenericType
    args []types.Type
    retType types.Type
    retAddr addr.Addr
    Scope *Scope
}

var curFunc *Func = nil

func GetCurFunc() *Func {
    return curFunc
}

func CreateFunc(name token.Token) Func {
    return Func{ name: name.Str, decPos: name.Pos }
}

func (f *Func) GetArgs() []types.Type {
    return f.args
}

func (f *Func) GetName() string {
    return f.name
}

func (f *Func) GetType() types.Type {
    // TODO
    return nil
}

func (f *Func) GetGeneric() *types.GenericType {
    return f.generic
}

func (f *Func) GetRetType() types.Type {
    return f.retType
}

func (f *Func) GetPos() token.Pos {
    return f.decPos
}

func (f *Func) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] TODO: func.go Addr()")
    os.Exit(1)
    return addr.Addr{}
}

func (f *Func) GetRetAddr() addr.Addr {
    return f.retAddr
}

func (f *Func) GetMangledName() string {
    if f.GetGeneric() != nil {
        return f.name + "$" + f.generic.CurUsedType.String()
    } else {
        return f.name
    }
}


func (f *Func) SetRetType(typ types.Type) {
    f.retType = typ
}

func (f *Func) SetRetAddr(addr addr.Addr) {
    f.retAddr = addr
}

func (f *Func) SetArgs(args []types.Type) {
    f.args = args
}

func (f *Func) SetGeneric(generic *types.GenericType) {
    f.generic = generic
}


func (f *Func) AddTypeToGeneric(typ types.Type) {
    for _,t := range f.generic.UsedTypes {
        if types.Equal(typ, t) { return }
    }

    f.generic.UsedTypes = append(f.generic.UsedTypes, typ)
}


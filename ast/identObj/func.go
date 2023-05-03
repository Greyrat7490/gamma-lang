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
    methodOf string
    isConst bool
}

var curFunc *Func = nil

func GetCurFunc() *Func {
    return curFunc
}

func CreateFunc(name token.Token, isConst bool) Func {
    return Func{ name: name.Str, decPos: name.Pos, isConst: isConst }
}

func CreateMethod(name token.Token, isConst bool, structName string) Func {
    return Func{ name: name.Str, decPos: name.Pos, isConst: isConst, methodOf: structName }
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

func (f *Func) GetMethodOf() string {
    return f.methodOf
}

func (f *Func) GetMangledName() string {
    name := f.name

    if f.methodOf != "" {
        name = f.methodOf + "." + name
    }

    if f.GetGeneric() != nil {
        name += "$" + f.generic.CurUsedType.String()
    }

    return name
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


func (f *Func) Equal(other *Func) bool {
    if f.name != other.name {
        return false
    }

    if f.generic != other.generic {
        return false
    }

    if f.retType != other.retType {
        return false
    }

    if len(f.args) != len(other.args) {
        return false
    }

    for i := range f.args {
        if !types.Equal(f.args[i], other.args[i]) {
            return false
        }
    }

    return true
}

func (f Func) String() string {
    generic := ""
    if f.generic != nil {
        generic = fmt.Sprintf("<%s>", f.generic)
    }

    ret := ""
    if f.retType != nil {
        ret = fmt.Sprintf(" -> %s", f.retType)
    }

    return fmt.Sprintf("%s%s(%v)%s", f.name, generic, f.args, ret)
}

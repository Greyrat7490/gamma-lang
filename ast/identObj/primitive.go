package identObj

import (
	"fmt"
	"gamma/token"
	"gamma/types"
	"gamma/types/addr"
	"os"
)

type Primitive struct {
    typ types.Type
    impls []Impl
    interfaces []string
}

func AddPrimitive(t types.Type) {
    globalScope.identObjs[t.String()] = &Primitive{ typ: t }
}

func (p *Primitive) GetName() string {
    return p.typ.String()
}

func (p *Primitive) GetPos() token.Pos {
    fmt.Fprintf(os.Stderr, "[ERROR] %s is buildin cannot get declaration position\n", p.typ.String())
    os.Exit(1)
    return token.Pos{}
}

func (p *Primitive) GetType() types.Type {
    return p.typ
}

func (p *Primitive) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] Cannot get the addr of a const (consts are not allocated anywhere)")
    os.Exit(1)
    return addr.Addr{}
}


func (p *Primitive) GetFuncNames() []string {
    funcs := []string{}

    for _,i := range p.impls {
        funcs = append(funcs, i.GetInterfaceFuncNames()...)
    }

    return funcs
}

func (p *Primitive) AddImpl(impl Impl) {
    p.impls = append(p.impls, impl)
    if impl.interface_ != nil {
        p.interfaces = append(p.interfaces, impl.interface_.name) 
    }
}

func (p *Primitive) HasInterface(name string) bool {
    for _,interfaceName := range p.interfaces {
        if interfaceName == name {
            return true
        }
    }
    return false
}

func (p *Primitive) GetFunc(name string) *Func {
    for _,i := range p.impls {
        f := i.GetFunc(name)
        if f != nil {
            return f
        }
    }

    return nil
}

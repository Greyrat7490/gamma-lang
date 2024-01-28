package identObj

import "gamma/types"

type Implementable struct {
    dstType types.Type
    impls []Impl
    interfaces []string
}

func (s *Implementable) AddImpl(impl Impl) {
    s.impls = append(s.impls, impl)
    if impl.interface_ != nil {
        s.interfaces = append(s.interfaces, impl.interface_.name) 
    }
}

func (s *Implementable) HasInterface(name string) bool {
    for _,interfaceName := range s.interfaces {
        if interfaceName == name {
            return true
        }
    }
    return false
}

func (s *Implementable) GetFunc(name string) *Func {
    for _,i := range s.impls {
        f := i.GetFunc(name)
        if f != nil {
            return f
        }
    }

    return nil
}

func (s *Implementable) GetFuncNames() []string {
    funcs := []string{}

    for _,i := range s.impls {
        funcs = append(funcs, i.GetInterfaceFuncNames()...)
    }

    return funcs
}

func HasFunc(t types.Type, name string) bool {
    if t == nil { return false }

    if t,ok := t.(types.InterfaceType); ok {
        return t.GetFunc(name) != nil
    }

    implObj := GetImplObj(t.String())
    return implObj != nil && implObj.GetFunc(name) != nil
}

func HasInterface(t types.Type, name string) bool {
    if t == nil { return false }

    if t,ok := t.(*types.GenericType); ok {
        if t.CurInsetType == nil {
            // TODO: generic guards (required interfaces list)
            return true
        }

        return HasInterface(t.CurInsetType, name)
    }

    if t,ok := t.(types.InterfaceType); ok {
        return t.Name == name
    }

    implObj := GetImplObj(t.String())
    return implObj != nil && implObj.HasInterface(name)
}

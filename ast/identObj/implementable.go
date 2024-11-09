package identObj

import "gamma/types"

type Implementations []*Implementable

type Implementable struct {
    impls []Impl
    interfaces []string
    subImpls Implementations
}

func reserve(arr Implementations, space uint64) Implementations {
    newArr := make(Implementations, space+50)
    copy(newArr, arr)
    return newArr
}

func createImplementable(impls Implementations, idx uint64) *Implementable {
    if impls[idx] == nil {
        impl := &Implementable{ impls: make([]Impl, 0, 5), interfaces: make([]string, 0, 5), subImpls: make(Implementations, 0, 5) }
        impls[idx] = impl
    }

    return impls[idx]
}

func (impls Implementations) get(idx uint64) *Implementable {
    if uint64(len(impls)) <= idx { 
        return nil
    }

    return impls[idx]
}

func (impls *Implementations) create(idx uint64) *Implementable {
    if uint64(len(*impls)) <= idx { 
        *impls = reserve(*impls, idx)
    }

    return createImplementable(*impls, idx)
}

func GetImplementable(t types.Type, createIfMissing bool) *Implementable {
    implId := t.GetImplID()

    var impl *Implementable = nil
    if createIfMissing {
        impl = globalScope.implObj.create(implId)
    } else {
        impl = globalScope.implObj.get(implId)
        if impl == nil { return nil }
    }

    switch t := t.(type) {
    case types.StructType:
        if t.GetInsetType() == nil { return impl }
        implId = t.GetInsetType().GetImplID()

    case types.EnumType:
        if t.GetInsetType() == nil { return impl }
        implId = t.GetInsetType().GetImplID()

    case types.InterfaceType:
        if t.Generic.SetType == nil { return impl }
        implId = t.Generic.SetType.GetImplID()

    case types.VecType:
        if t.BaseType == nil { return impl }
        implId = t.BaseType.GetImplID()

    case types.PtrType:
        if t.BaseType == nil { return impl }
        implId = t.BaseType.GetImplID()

    case types.ArrType:
        if t.BaseType == nil { return impl }
        implId = t.BaseType.GetImplID()

    case types.InferType:
        return GetImplementable(t.DefaultType, createIfMissing) 

    default:
        return impl
    }

    if createIfMissing {
        return impl.subImpls.create(implId)
    } else {
        return impl.subImpls.get(implId)
    }
}

func (s *Implementable) AddImpl(impl Impl) {
    s.impls = append(s.impls, impl)
    if impl.interfaceType != nil {
        s.interfaces = append(s.interfaces, impl.interfaceType.Name) 
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
        if f != nil { return f }
    }

    return nil
}

func (s *Implementable) GetImplByFnName(fnName string) *Impl {
    for _,i := range s.impls {
        if i.GetFunc(fnName) != nil {
            return &i
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

    switch t := t.(type) {
    case *types.GenericType:
        return HasFunc(t.Guard, name)
    case types.GenericType:
        return HasFunc(t.Guard, name)
    case types.InterfaceType:
        return t.GetFunc(name) != nil
    }

    implObj := GetImplementable(t, false)
    return implObj != nil && implObj.GetFunc(name) != nil
}

func HasInterface(t types.Type, name string) bool {
    if t == nil { return false }

    switch t := t.(type) {
    case *types.GenericType:
        return HasInterface(t.Guard, name) 
    case types.GenericType:
        return HasInterface(t.Guard, name) 
    case types.InterfaceType:
        return t.Name == name
    }

    implObj := GetImplementable(t, false)
    return implObj != nil && implObj.HasInterface(name)
}

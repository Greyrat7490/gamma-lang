package types

var genericInsetTypes []Type = make([]Type, 0, 20)

func CreateGeneric(name string) GenericType {
    genericInsetTypes = append(genericInsetTypes, nil)
    return GenericType{ Name: name, Idx: uint64(len(genericInsetTypes)-1) }
}

func UpdateInsetType(t Type) {
    switch t := t.(type) {
    case PtrType:
        UpdateInsetType(t.BaseType)

    case ArrType:
        UpdateInsetType(t.BaseType)

    case VecType:
        UpdateInsetType(t.BaseType)

    case GenericType:
        SetCurInsetType(t, t.SetType)
    case *GenericType:
        SetCurInsetType(t, t.SetType)
    }
}

func SetCurInsetType(t Type, insetType Type) {
    if insetType == nil { return }
    
    switch t := t.(type) {
    case PtrType:
        SetCurInsetType(t.BaseType, insetType)

    case ArrType:
        SetCurInsetType(t.BaseType, insetType)

    case VecType:
        SetCurInsetType(t.BaseType, insetType)

    case GenericType:
        if t.Idx < uint64(len(genericInsetTypes)) {
            genericInsetTypes[t.Idx] = insetType 
        }
    case *GenericType:
        if t.Idx < uint64(len(genericInsetTypes)) {
            genericInsetTypes[t.Idx] = insetType 
        }
    }
}

func ResolveGeneric(t Type) Type {
    switch t := t.(type) {
    case PtrType:
        t.BaseType = ResolveGeneric(t.BaseType)
        return t

    case ArrType:
        t.BaseType = ResolveGeneric(t.BaseType)
        return t

    case VecType:
        t.BaseType = ResolveGeneric(t.BaseType)
        return t

    case GenericType:
        if t.SetType != nil {
            return t.SetType
        }
        
        if t.Idx < uint64(len(genericInsetTypes)) && genericInsetTypes[t.Idx] != nil {
            return genericInsetTypes[t.Idx]
        }

        return nil
    case *GenericType:
        if t.SetType != nil {
            return t.SetType
        }

        if t.Idx < uint64(len(genericInsetTypes)) && genericInsetTypes[t.Idx] != nil {
            return genericInsetTypes[t.Idx]
        }

        return nil
    }

    return t
}

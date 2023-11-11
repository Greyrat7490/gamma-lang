package identObj

import "gamma/types"

var curGeneric *types.GenericType = nil

func GetGeneric(name string) *types.GenericType {
    if curGeneric != nil && curGeneric.Name == name {
        return curGeneric
    }

    return nil
}

func SetGeneric(t *types.GenericType) {
    curGeneric = t
}

func AddTypeToGeneric(generic *types.GenericType, typ types.Type) {
    for _,t := range generic.UsedTypes {
        if types.Equal(typ, t) { return }
    }

    generic.UsedTypes = append(generic.UsedTypes, typ)
}

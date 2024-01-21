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

func UnsetGeneric() {
    curGeneric = nil
}

func AddTypeToGeneric(generic *types.GenericType, typ types.Type) {
    for _,t := range generic.UsedInsetTypes {
        if types.Equal(typ, t) { return }
    }

    generic.UsedInsetTypes = append(generic.UsedInsetTypes, typ)
}

package identObj

import (
	"fmt"
	"gamma/token"
	"gamma/types"
	"gamma/types/addr"
	"os"
)

type Generic struct {
    decPos token.Pos
    Typ types.GenericType
    UsedInsetTypes []types.Type
}

func CreateGeneric(name token.Token, guardType types.InterfaceType) Generic {
    return Generic{ Typ: types.CreateGeneric(name.Str, guardType), decPos: name.Pos }
}

func (g *Generic) GetName() string {
    return g.Typ.Name
}

func (g *Generic) GetPos() token.Pos {
    return g.decPos
}

func (g *Generic) GetType() types.Type {
    return &g.Typ
}

func (g *Generic) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] (internal) Cannot get the addr of Generic (Generic are not allocated anywhere)")
    os.Exit(1)
    return addr.Addr{}
}

func AddTypeToGeneric(generic *Generic, typ types.Type) {
    if typ.GetKind() == types.Infer { return }

    for _,t := range generic.UsedInsetTypes {
        if types.Equal(typ, t) { return }
    }

    generic.UsedInsetTypes = append(generic.UsedInsetTypes, typ)
}

func (g *Generic) RemoveDuplTypes() {
    for i := range g.UsedInsetTypes {
        for j := i+1; j < len(g.UsedInsetTypes); j++ {
            if types.Equal(g.UsedInsetTypes[i], g.UsedInsetTypes[j]) && g.UsedInsetTypes[i].Size() == g.UsedInsetTypes[j].Size() {
                g.UsedInsetTypes[j] = g.UsedInsetTypes[len(g.UsedInsetTypes)-1]
                g.UsedInsetTypes = g.UsedInsetTypes[:len(g.UsedInsetTypes)-1]
            }
        }
    }
}

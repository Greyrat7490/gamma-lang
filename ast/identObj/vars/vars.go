package vars

import (
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type Var interface {
    Addr() addr.Addr
    GetType() types.Type
    GetName() string
    GetPos() token.Pos
    String() string
    ResolveType(t types.Type, useDefault bool)
}

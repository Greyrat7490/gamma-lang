package identObj

import (
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type IdentObj interface {
    GetName() string
    GetType() types.Type
    GetPos() token.Pos
    Addr() addr.Addr
}

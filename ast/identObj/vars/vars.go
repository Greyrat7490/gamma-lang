package vars

import (
    "gamma/token"
    "gamma/types"
)

type Var interface {
    Addr(field uint) string
    OffsetedAddr(offset int) string

    GetType() types.Type
    GetName() string
    GetPos() token.Pos
    String() string
}

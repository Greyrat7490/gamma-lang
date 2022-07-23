package consts

import (
    "gorec/token"
    "gorec/types"
)

type Const struct {
    Name token.Token
    Type types.Type
    Val token.Token
}

func (c *Const) Define(val token.Token) {
    c.Val = val
}

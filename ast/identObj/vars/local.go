package vars

import (
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type LocalVar struct {
    decPos token.Pos
    name string
    typ types.Type
    addr addr.Addr
}

func CreateLocal(name token.Token, t types.Type) LocalVar {
    return LocalVar{ name: name.Str, decPos: name.Pos, typ: t }
}

func (v *LocalVar) String() string {
    return fmt.Sprintf("{%s %v}", v.name, v.typ)
}

func (v *LocalVar) GetName() string {
    return v.name
}

func (v *LocalVar) GetPos() token.Pos {
    return v.decPos
}

func (v *LocalVar) GetType() types.Type {
    return v.typ
}

func (v *LocalVar) Addr() addr.Addr {
    return v.addr
}

func (v *LocalVar) ResolveType(t types.Type) {
    if v.typ.GetKind() == types.Infer {
        if inferType,ok := t.(types.InferType); ok {
            v.typ = inferType.DefaultType
        } else {
            v.typ = t
        }
    }
}

func (v *LocalVar) SetOffset(frameSize uint, fromStack bool) {
    offset := int64(0)
    if fromStack {
        offset = int64(types.Ptr_Size + frameSize + 7) & ^7
    } else {
        offset = -int64(frameSize + v.typ.Size())
    }

    v.addr = addr.Addr{ BaseAddr: "rbp", Offset: offset }
}

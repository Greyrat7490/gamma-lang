package fn

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
)

type Func struct {
    decPos token.Pos
    name string
    args []types.Type
    retType types.Type
    frameSize uint
}

func CreateFunc(name token.Token, args []types.Type, retType types.Type) Func {
    // frameSize = 1 -> invalid value
    return Func{ name: name.Str, decPos: name.Pos, args: args, retType: retType, frameSize: 1 }
}

func (f *Func) GetArgs() []types.Type {
    return f.args
}

func (f *Func) GetName() string {
    return f.name
}

func (f *Func) GetType() types.Type {
    // TODO
    return nil
}

func (f *Func) GetRetType() types.Type {
    return f.retType
}

func (f *Func) GetPos() token.Pos {
    return f.decPos
}

func (f *Func) SetFrameSize(size uint) {
    if f.frameSize != 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] setting the frameSize of a function again is not allowed")
        os.Exit(1)
    }

    // size has to be the multiple of 16byte
    f.frameSize = (size + 15) & ^uint(15)
}

func (f *Func) GetFrameSize() uint {
    return f.frameSize
}

func (f *Func) Addr() addr.Addr {
    fmt.Fprintln(os.Stderr, "[ERROR] TODO: func.go Addr()")
    os.Exit(1)
    return addr.Addr{}
}

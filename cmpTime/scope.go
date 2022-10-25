package cmpTime

import (
    "os"
    "fmt"
    "reflect"
    "strconv"
    "unsafe"
    "encoding/binary"
    "gamma/token"
    "gamma/types"
    "gamma/cmpTime/constVal"
)

var curScope *scope = nil
var stack []byte = nil

type scope struct {
    consts map[string]constVal.ConstVal
    parent *scope
}

func initStack(framesize uint) {
    stack = make([]uint8, framesize)
}

func clearStack() {
    stack = nil
}

func startScope() {
    curScope = &scope{ parent: curScope, consts: make(map[string]constVal.ConstVal) }
}

func endScope() {
    curScope = curScope.parent
}

func inConstEnv() bool {
    return curScope != nil
}

func defVar(name string, addr string, t types.Type, pos token.Pos, val constVal.ConstVal) {
    if _,ok := curScope.consts[name]; ok {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is already declared in this scope\n", name)
        fmt.Fprintln(os.Stderr, "\t", pos.At())
        os.Exit(1)
    }

    curScope.consts[name] = val

    if idx := getStackIdx(addr, t); idx != -1 {
        writeStack(idx, t.Size(), val)
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] invalid addr %s (with type %v)\n", addr, t)
        fmt.Fprintln(os.Stderr, "\t", pos.At())
        os.Exit(1)
    }
}

func defConst(name string, pos token.Pos, val constVal.ConstVal) {
    if _,ok := curScope.consts[name]; ok {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is already declared in this scope\n", name)
        fmt.Fprintln(os.Stderr, "\t", pos.At())
        os.Exit(1)
    }

    curScope.consts[name] = val
}

func setVar(name string, addr string, t types.Type, pos token.Pos, val constVal.ConstVal) {
    cur := curScope
    for cur != nil {
        if _,ok := cur.consts[name]; ok {
            cur.consts[name] = val
            if idx := getStackIdx(addr, t); idx != -1 {
                writeStack(idx, t.Size(), val)
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] invalid addr %s (with type %v)\n", addr, t)
                fmt.Fprintln(os.Stderr, "\t", pos.At())
                os.Exit(1)
            }
            return
        } else {
            cur = cur.parent
        }
    }

    // TODO: better error message (check if ident of global or non const local var)
    fmt.Fprintf(os.Stderr, "[ERROR] %s is not declared\n", name)
    fmt.Fprintln(os.Stderr, "\t", pos.At())
    os.Exit(1)
}

func getVal(name string, pos token.Pos) constVal.ConstVal {
    cur := curScope
    for cur != nil {
        if c,ok := cur.consts[name]; ok {
            return c
        } else {
            cur = cur.parent
        }
    }

    return nil
}

func getValFromStack(addr string, t types.Type) constVal.ConstVal {
    if idx := getStackIdx(addr, t); idx != -1 {
        return readStack(idx, t)
    } else {
        return nil
    }
}

func getOffset(addr string) int {
    if len(addr) > 4 {
        if offset,err := strconv.Atoi(addr[4:]); err == nil {
            return offset
        }
    }

    return 0
}

func getStackIdx(addr string, t types.Type) int {
    offset := getOffset(addr)

    if idx := offset - int(t.Size()); idx >= 0 && idx < len(stack) {
        return idx
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %s (of type %v) is outside of the stack (size: %d)\n", addr, t, len(stack))
        os.Exit(1)
        return -1
    }
}

func readStack(idx int, t types.Type) constVal.ConstVal {
    switch t.GetKind() {
    case types.Int:
        var c int64
        switch t.Size() {
        case 1:
            c = int64(stack[idx])
        case 2:
            c = int64(getByteOrder().Uint16(stack[idx:]))
        case 4:
            c = int64(getByteOrder().Uint32(stack[idx:]))
        default:
            c = int64(getByteOrder().Uint64(stack[idx:]))
        }
        return (*constVal.IntConst)(&c)

    case types.Uint:
        var c uint64
        switch t.Size() {
        case 1:
            c = uint64(stack[idx])
        case 2:
            c = uint64(getByteOrder().Uint16(stack[idx:]))
        case 4:
            c = uint64(getByteOrder().Uint32(stack[idx:]))
        default:
            c = getByteOrder().Uint64(stack[idx:])
        }
        return (*constVal.UintConst)(&c)

    case types.Bool:
        if stack[idx] == 0 {
            b := constVal.BoolConst(false)
            return &b
        } else {
            b := constVal.BoolConst(true)
            return &b
        }

    case types.Char:
        return (*constVal.CharConst)(&stack[idx])

    case types.Ptr:
        offset := getByteOrder().Uint64(stack[idx:])
        c := constVal.PtrConst{ Local: true, Addr: fmt.Sprintf("rbp-%d", offset) }
        return &c

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] reading a %v from stack is not supported yet\n", t)
        os.Exit(1)
        return nil
    }
}

func writeStack(idx int, size uint, val constVal.ConstVal) {
    switch c := val.(type) {
    case *constVal.IntConst:
        switch size {
        case 1:
            stack[idx] = byte(*c)
        case 2:
            getByteOrder().PutUint16(stack[idx:], uint16(*c))
        case 4:
            getByteOrder().PutUint32(stack[idx:], uint32(*c))
        default:
            getByteOrder().PutUint64(stack[idx:], uint64(*c))
        }

    case *constVal.UintConst:
        switch size {
        case 1:
            stack[idx] = byte(*c)
        case 2:
            getByteOrder().PutUint16(stack[idx:], uint16(*c))
        case 4:
            getByteOrder().PutUint32(stack[idx:], uint32(*c))
        default:
            getByteOrder().PutUint64(stack[idx:], uint64(*c))
        }

    case *constVal.BoolConst:
        if bool(*c) {
            stack[idx] = 1
        } else {
            stack[idx] = 0
        }

    case *constVal.CharConst:
        stack[idx] = byte(*c)

    case *constVal.PtrConst:
        if offset := getOffset(c.Addr); offset != 0 {
            getByteOrder().PutUint64(stack[idx:], uint64(offset))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] invalid addr %s\n", c.Addr)
            os.Exit(1)
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] writing a %v from stack is not supported yet\n", reflect.TypeOf(val))
        os.Exit(1)
    }
}

func getByteOrder() binary.ByteOrder {
    buf := [2]byte{}
    *((*uint16)(unsafe.Pointer(&buf[0]))) = uint16(0x0001)

    // opposite of native endianness (stack grows from top to bottom)
    if buf[0] == 0x01 {
        return binary.BigEndian
    } else {
        return binary.LittleEndian
    }
}

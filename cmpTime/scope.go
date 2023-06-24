package cmpTime

import (
    "os"
    "fmt"
    "unsafe"
    "reflect"
    "encoding/binary"
    "gamma/token"
    "gamma/types"
    "gamma/types/addr"
    "gamma/types/array"
    "gamma/cmpTime/constVal"
)

var curScope *scope = nil

type scope struct {
    consts map[string]constVal.ConstVal
    vars map[string]varInfo
    stack []byte
    parent *scope
}

type varInfo struct {
    stackIdx uint
    typ types.Type
}

func startScope(framesize uint) {
    curScope = &scope{ parent: curScope, consts: make(map[string]constVal.ConstVal), vars: make(map[string]varInfo), stack: make([]uint8, framesize) }
}

func endScope() {
    curScope = curScope.parent
}

func inConstEnv() bool {
    return curScope != nil
}

func checkNameTaken(name string, pos token.Pos) {
    if _,ok := curScope.consts[name]; ok {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is already declared in this scope\n", name)
        fmt.Fprintln(os.Stderr, "\t", pos.At())
        os.Exit(1)
    }
    if _,ok := curScope.vars[name]; ok {
        fmt.Fprintf(os.Stderr, "[ERROR] %s is already declared in this scope\n", name)
        fmt.Fprintln(os.Stderr, "\t", pos.At())
        os.Exit(1)
    }
}

func defVar(name string, addr addr.Addr, t types.Type, pos token.Pos, val constVal.ConstVal) {
    checkNameTaken(name, pos)

    idx := getStackIdx(addr, t)
    curScope.vars[name] = varInfo{ stackIdx: idx, typ: t }
    writeStack(idx, t, val)
}

func defConst(name string, pos token.Pos, val constVal.ConstVal) {
    checkNameTaken(name, pos)

    curScope.consts[name] = val
}

func setVar(name string, t types.Type, pos token.Pos, val constVal.ConstVal) {
    cur := curScope
    for cur != nil {
        if v,ok := cur.vars[name]; ok {
            writeStack(v.stackIdx, t, val)
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

func setVarAddr(addr addr.Addr, t types.Type, val constVal.ConstVal) {
    writeStack(getStackIdx(addr, t), t, val)
}

func setVarField(name string, offset uint, t types.Type, pos token.Pos, val constVal.ConstVal) {
    cur := curScope
    for cur != nil {
        if v,ok := cur.vars[name]; ok {
            writeStack(v.stackIdx + offset, t, val)
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
        } else if c,ok := cur.vars[name]; ok {
            return readStack(c.stackIdx, c.typ)
        } else {
            cur = cur.parent
        }
    }

    return nil
}

func getValAddr(addr addr.Addr, t types.Type) constVal.ConstVal {
    return readStack(getStackIdx(addr, t), t)
}

func getStackIdx(addr addr.Addr, t types.Type) uint {
    if idx := -addr.Offset - int64(t.Size()); idx >= 0 && idx < int64(len(curScope.stack)) {
        return uint(idx)
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %s (of type %v) is outside of the stack (size: %d)\n", addr, t, len(curScope.stack))
        os.Exit(1)
        return 0
    }
}

func readStack(idx uint, t types.Type) constVal.ConstVal {
    switch t := t.(type) {
    case types.IntType:
        var c int64
        switch t.Size() {
        case 1:
            c = int64(int8(curScope.stack[idx]))     // to sign extend
        case 2:
            c = int64(int16(getByteOrder().Uint16(curScope.stack[idx:])))
        case 4:
            c = int64(int32(getByteOrder().Uint32(curScope.stack[idx:])))
        default:
            c = int64(getByteOrder().Uint64(curScope.stack[idx:]))
        }
        return (*constVal.IntConst)(&c)

    case types.UintType:
        var c uint64
        switch t.Size() {
        case 1:
            c = uint64(curScope.stack[idx])
        case 2:
            c = uint64(getByteOrder().Uint16(curScope.stack[idx:]))
        case 4:
            c = uint64(getByteOrder().Uint32(curScope.stack[idx:]))
        default:
            c = getByteOrder().Uint64(curScope.stack[idx:])
        }
        return (*constVal.UintConst)(&c)

    case types.BoolType:
        if curScope.stack[idx] == 0 {
            b := constVal.BoolConst(false)
            return &b
        } else {
            b := constVal.BoolConst(true)
            return &b
        }

    case types.CharType:
        return (*constVal.CharConst)(&curScope.stack[idx])

    case types.PtrType:
        offset := getByteOrder().Uint64(curScope.stack[idx:])
        c := constVal.PtrConst{ Local: true, Addr: addr.Addr{ BaseAddr: "rbp", Offset: int64(offset) } }
        return &c

    case types.ArrType:
        idx := getByteOrder().Uint64(curScope.stack[idx:])
        return &constVal.ArrConst{ Idx: idx, Elems: array.GetValues(idx), Type: t }

    case types.StructType:
        c := constVal.StructConst{ Fields: make([]constVal.ConstVal, len(t.Types)) }
        for i,t := range t.Types {
            c.Fields[i] = readStack(idx, t)
            idx += t.Size()
        }
        return &c

    case types.EnumType:
        idVal := readStack(idx, t.IdType)
        idx += t.IdType.Size()

        id := uint64(0)
        switch c := idVal.(type) {
        case *constVal.UintConst:
            id = uint64(*c)
        case *constVal.IntConst:
            id = uint64(*c)
        }

        elemType := t.GetTypeWithId(id)
        if elemType  != nil {
            elem := readStack(idx, elemType)
            idx += t.Size()
            return &constVal.EnumConst{ Id: id, Type: t, ElemType: elemType, Elem: elem }
        } else {
            return &constVal.EnumConst{ Id: id, Type: t, ElemType: elemType }
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] reading %v from the const stack is not supported yet\n", t)
        os.Exit(1)
        return nil
    }
}

func writeStack(idx uint, typ types.Type, val constVal.ConstVal) {
    switch c := val.(type) {
    case *constVal.IntConst:
        switch typ.Size() {
        case 1:
            curScope.stack[idx] = byte(*c)
        case 2:
            getByteOrder().PutUint16(curScope.stack[idx:], uint16(*c))
        case 4:
            getByteOrder().PutUint32(curScope.stack[idx:], uint32(*c))
        default:
            getByteOrder().PutUint64(curScope.stack[idx:], uint64(*c))
        }

    case *constVal.UintConst:
        switch typ.Size() {
        case 1:
            curScope.stack[idx] = byte(*c)
        case 2:
            getByteOrder().PutUint16(curScope.stack[idx:], uint16(*c))
        case 4:
            getByteOrder().PutUint32(curScope.stack[idx:], uint32(*c))
        default:
            getByteOrder().PutUint64(curScope.stack[idx:], uint64(*c))
        }

    case *constVal.BoolConst:
        if bool(*c) {
            curScope.stack[idx] = 1
        } else {
            curScope.stack[idx] = 0
        }

    case *constVal.CharConst:
        curScope.stack[idx] = byte(*c)

    case *constVal.PtrConst:
        getByteOrder().PutUint64(curScope.stack[idx:], uint64(c.Addr.Offset))

    case *constVal.ArrConst:
        getByteOrder().PutUint64(curScope.stack[idx:], c.Idx)

    case *constVal.StructConst:
        for i,field := range c.Fields {
            t := typ.(types.StructType).Types[i]
            writeStack(idx, t, field)
            idx += t.Size()
        }

    case *constVal.EnumConst:
        id := constVal.UintConst(c.Id)
        writeStack(idx, c.Type.IdType, &id)
        idx += c.Type.IdType.Size()
        if c.ElemType != nil {
            writeStack(idx, c.ElemType, c.Elem)
            idx += c.ElemType.Size()
        }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] writing %v to the const stack is not supported yet\n", reflect.TypeOf(val))
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

package asm

import (
    "os"
    "fmt"
    "gamma/token"
    "gamma/cmpTime/constVal"
)

// results in A-register
func BinaryOp(file *os.File, opType token.TokenType, src string, size uint, signed bool) {
    switch opType {
    case token.Eql:
        Eql(file, GetReg(RegA, size), src)
    case token.Neq:
        Neq(file, GetReg(RegA, size), src)
    case token.Lss:
        Lss(file, GetReg(RegA, size), src)
    case token.Grt:
        Grt(file, GetReg(RegA, size), src)
    case token.Leq:
        Leq(file, GetReg(RegA, size), src)
    case token.Geq:
        Geq(file, GetReg(RegA, size), src)

    case token.Plus:
        Add(file, src, size)
    case token.Minus:
        Sub(file, src, size)
    case token.Mul:
        Mul(file, src, size, signed)
    case token.Div:
        Div(file, src, size, signed)
    case token.Mod:
        Mod(file, src, size, signed)

    case token.Shl:
        Shl(file, src, size)
    case token.Shr:
        Shr(file, src, size)
    case token.Amp:
        And(file, src, size)
    case token.BitOr:
        Or(file, src, size)
    case token.Xor:
        Xor(file, src, size)

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %v\n", opType)
        os.Exit(1)
    }
}

func BinaryOpReg(file *os.File, opType token.TokenType, reg RegGroup, size uint, signed bool) {
    BinaryOp(file, opType, GetReg(reg, size), size, signed)
}

func BinaryOpEvalBools(op token.Token, lhs bool, rhs bool) constVal.ConstVal {
    var b1 int64 = 0
    var b2 int64 = 0

    if lhs {
        b1 = 1
    }

    if rhs {
        b2 = 1
    }

    return BinaryOpEvalInts(op, b1, b2)
}

func BinaryOpEvalUints(op token.Token, lhs uint64, rhs uint64) constVal.ConstVal {
    return BinaryOpEvalInts(op, int64(lhs), int64(rhs))
}

func BinaryOpEvalInts(op token.Token, lhs int64, rhs int64) constVal.ConstVal {
    switch op.Type {
    case token.Eql:
        if lhs == rhs {
            c := constVal.BoolConst(true)
            return &c
        } else {
            c := constVal.BoolConst(false)
            return &c
        }

    case token.Neq:
        if lhs != rhs {
            c := constVal.BoolConst(true)
            return &c
        } else {
            c := constVal.BoolConst(false)
            return &c
        }

    case token.Lss:
        if lhs < rhs {
            c := constVal.BoolConst(true)
            return &c
        } else {
            c := constVal.BoolConst(false)
            return &c
        }

    case token.Grt:
        if lhs > rhs {
            c := constVal.BoolConst(true)
            return &c
        } else {
            c := constVal.BoolConst(false)
            return &c
        }

    case token.Leq:
        if lhs <= rhs {
            c := constVal.BoolConst(true)
            return &c
        } else {
            c := constVal.BoolConst(false)
            return &c
        }

    case token.Geq:
        if lhs >= rhs {
            c := constVal.BoolConst(true)
            return &c
        } else {
            c := constVal.BoolConst(false)
            return &c
        }

    case token.Plus:
        c := constVal.IntConst(lhs + rhs)
        return &c

    case token.Minus:
        c := constVal.IntConst(lhs - rhs)
        return &c

    case token.Mul:
        c := constVal.IntConst(lhs * rhs)
        return &c

    case token.Div:
        c := constVal.IntConst(lhs / rhs)
        return &c

    case token.Mod:
        c := constVal.IntConst(lhs % rhs)
        return &c


    case token.Shl:
        c := constVal.IntConst(lhs << rhs)
        return &c

    case token.Shr:
        c := constVal.IntConst(lhs >> rhs)
        return &c

    case token.Amp:
        c := constVal.IntConst(lhs & rhs)
        return &c

    case token.BitOr:
        c := constVal.IntConst(lhs | rhs)
        return &c

    case token.Xor:
        c := constVal.IntConst(lhs ^ rhs)
        return &c

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %v\n", op)
        os.Exit(1)
        return nil
    }
}

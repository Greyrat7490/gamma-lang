package asm

import (
    "os"
    "fmt"
    "bufio"
    "gamma/token"
    "gamma/types"
    "gamma/types/str"
    "gamma/cmpTime/constVal"
)

// results in A-register
func BinaryOpStrsLit(file *bufio.Writer, opType token.TokenType, rhsIdx uint64) {
    MovRegVal(file, RegB, types.Ptr_Size, fmt.Sprintf("_str%d", rhsIdx))
    MovRegVal(file, RegC, types.U32_Size, fmt.Sprint(str.GetSize(rhsIdx)))
    BinaryOpStrs(file, opType)
}

func BinaryOpStrs(file *bufio.Writer, opType token.TokenType) {
    switch opType {
    case token.Eql:
        file.WriteString("call _str_cmp\n")
    case token.Neq:
        file.WriteString("call _str_cmp\n")
        file.WriteString("xor al, 1\n")
    default:
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected binary operator for str expected (== or !=)")
        os.Exit(1)
    }
}

func BinaryOp(file *bufio.Writer, opType token.TokenType, src string, t types.Type) {
    switch t.GetKind() {
    case types.Char:
        switch opType {
        case token.Eql:
            Eql(file, GetReg(RegA, t.Size()), src)
        case token.Neq:
            Neq(file, GetReg(RegA, t.Size()), src)
        default:
            fmt.Fprintln(os.Stderr, "[ERROR] unexpected binary operator for char expected (== or !=)")
            os.Exit(1)
        }

    case types.Bool:
        switch opType {
        case token.Eql:
            Eql(file, GetReg(RegA, t.Size()), src)
        case token.Neq:
            Neq(file, GetReg(RegA, t.Size()), src)
        case token.Lss:
            Lss(file, GetReg(RegA, t.Size()), src)
        case token.Grt:
            Grt(file, GetReg(RegA, t.Size()), src)
        case token.Leq:
            Leq(file, GetReg(RegA, t.Size()), src)
        case token.Geq:
            Geq(file, GetReg(RegA, t.Size()), src)
        default:
            fmt.Fprintln(os.Stderr, "[ERROR] unexpected binary operator for bool")
            // TODO print pos
            // TODO print allowed ops
            os.Exit(1)
        }

    case types.Ptr:
        switch opType {
        case token.Eql:
            Eql(file, GetReg(RegA, t.Size()), src)
        case token.Neq:
            Neq(file, GetReg(RegA, t.Size()), src)
        case token.Lss:
            Lss(file, GetReg(RegA, t.Size()), src)
        case token.Grt:
            Grt(file, GetReg(RegA, t.Size()), src)
        case token.Leq:
            Leq(file, GetReg(RegA, t.Size()), src)
        case token.Geq:
            Geq(file, GetReg(RegA, t.Size()), src)

        case token.Plus:
            Add(file, src, t.Size())
        case token.Minus:
            Sub(file, src, t.Size())
        default:
            fmt.Fprintln(os.Stderr, "[ERROR] unexpected binary operator for ptr")
            os.Exit(1)
        }


    case types.Uint, types.Int:
        switch opType {
        case token.Eql:
            Eql(file, GetReg(RegA, t.Size()), src)
        case token.Neq:
            Neq(file, GetReg(RegA, t.Size()), src)
        case token.Lss:
            Lss(file, GetReg(RegA, t.Size()), src)
        case token.Grt:
            Grt(file, GetReg(RegA, t.Size()), src)
        case token.Leq:
            Leq(file, GetReg(RegA, t.Size()), src)
        case token.Geq:
            Geq(file, GetReg(RegA, t.Size()), src)

        case token.Plus:
            Add(file, src, t.Size())
        case token.Minus:
            Sub(file, src, t.Size())
        case token.Mul:
            Mul(file, src, t.Size(), t.GetKind() == types.Int)
        case token.Div:
            Div(file, src, t.Size(), t.GetKind() == types.Int)
        case token.Mod:
            Mod(file, src, t.Size(), t.GetKind() == types.Int)

        case token.Shl:
            Shl(file, src, t.Size())
        case token.Shr:
            Shr(file, src, t.Size())
        case token.Amp:
            And(file, src, t.Size())
        case token.BitOr:
            Or(file, src, t.Size())
        case token.Xor:
            Xor(file, src, t.Size())

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %v\n", opType)
            os.Exit(1)
        }
    case types.Str:
        fmt.Fprintln(os.Stderr, "[ERROR] (internal) BinaryOpStrs shoulb be called instead of BinaryOp")
        os.Exit(1)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected type for binary operation %v\n", t)
        os.Exit(1)
    }
}

func BinaryOpReg(file *bufio.Writer, opType token.TokenType, reg RegGroup, t types.Type) {
    BinaryOp(file, opType, GetReg(reg, t.Size()), t)
}

func BinaryOpEvalStrs(op token.Token, lhsIdx uint64, rhsIdx uint64) constVal.ConstVal {
    switch op.Type {
    case token.Eql:
        c := constVal.BoolConst(str.CmpStrLits(lhsIdx, rhsIdx))
        return &c
    case token.Neq:
        c := constVal.BoolConst(!str.CmpStrLits(lhsIdx, rhsIdx))
        return &c
    default:
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected binary operator for str expected (== or !=)")
        os.Exit(1)
        return nil
    }
}

func BinaryOpEvalBools(op token.Token, lhs bool, rhs bool) constVal.ConstVal {
    var b1 int64 = 0
    var b2 int64 = 0

    if lhs { b1 = 1 }
    if rhs { b2 = 1 }

    return BinaryOpEvalInts(op, b1, b2)
}

func BinaryOpEvalUints(op token.Token, lhs uint64, rhs uint64) constVal.ConstVal {
    return BinaryOpEvalInts(op, int64(lhs), int64(rhs))
}

func BinaryOpEvalInts(op token.Token, lhs int64, rhs int64) constVal.ConstVal {
    switch op.Type {
    case token.Eql:
        c := constVal.BoolConst(lhs == rhs)
        return &c

    case token.Neq:
        c := constVal.BoolConst(lhs != rhs)
        return &c

    case token.Lss:
        c := constVal.BoolConst(lhs < rhs)
        return &c

    case token.Grt:
        c := constVal.BoolConst(lhs > rhs)
        return &c

    case token.Leq:
        c := constVal.BoolConst(lhs <= rhs)
        return &c

    case token.Geq:
        c := constVal.BoolConst(lhs >= rhs)
        return &c

    case token.And:
        c := constVal.BoolConst(lhs == 1 && rhs == 1)
        return &c

    case token.Or:
        c := constVal.BoolConst(lhs == 1 || rhs == 1)
        return &c

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

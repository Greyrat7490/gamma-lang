package asm

import (
    "os"
    "fmt"
    "strconv"
    "gamma/token"
)

// results in A-register
func BinaryOp(file *os.File, opType token.TokenType, src string, size uint) {
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
        Mul(file, src, size)
    case token.Div:
        Div(file, src, size)
    case token.Mod:
        Mod(file, src, size)

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

func BinaryOpReg(file *os.File, opType token.TokenType, reg RegGroup, size uint) {
    BinaryOp(file, opType, GetReg(reg, size), size)
}

func BinaryOpVals(op token.Token, lhs token.Token, rhs token.Token) token.Token {
    switch op.Type {
    case token.Eql:
        if lhs.Str == rhs.Str {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Neq:
        if lhs.Str != rhs.Str {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }
    default:
        if lhs.Type != token.Number {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a number but got %v\n", lhs.Type)
            fmt.Fprintln(os.Stderr, "\t" + lhs.Pos.At())
            os.Exit(1)
        }
        if rhs.Type != token.Number {
            fmt.Fprintf(os.Stderr, "[ERROR] expected a number but got %v\n", rhs.Type)
            fmt.Fprintln(os.Stderr, "\t" + rhs.Pos.At())
            os.Exit(1)
        }
        tooBig, i1, i2 := toInts(lhs, rhs)

        if !tooBig {
            return evalInts(op, i1, i2)
        } else {
            u1, u2 := toUints(lhs, rhs)
            return evalUints(op, u1, u2)
        }
    }
}

func evalInts(op token.Token, i1 int64, i2 int64) token.Token {
    switch op.Type {
    case token.Lss:
        if i1 < i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Grt:
        if i1 > i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Leq:
        if i1 <= i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Geq:
        if i1 >= i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Plus:
        return token.Token{ Str: fmt.Sprint(i1 + i2), Type: token.Number, Pos: op.Pos }

    case token.Minus:
        return token.Token{ Str: fmt.Sprint(i1 - i2), Type: token.Number, Pos: op.Pos }

    case token.Mul:
        return token.Token{ Str: fmt.Sprint(i1 * i2), Type: token.Number, Pos: op.Pos }

    case token.Div:
        return token.Token{ Str: fmt.Sprint(i1 / i2), Type: token.Number, Pos: op.Pos }

    case token.Mod:
        return token.Token{ Str: fmt.Sprint(i1 % i2), Type: token.Number, Pos: op.Pos }


    case token.Shl:
        return token.Token{ Str: fmt.Sprint(i1 << i2), Type: token.Number, Pos: op.Pos }

    case token.Shr:
        return token.Token{ Str: fmt.Sprint(i1 >> i2), Type: token.Number, Pos: op.Pos }

    case token.Amp:
        return token.Token{ Str: fmt.Sprint(i1 & i2), Type: token.Number, Pos: op.Pos }

    case token.BitOr:
        return token.Token{ Str: fmt.Sprint(i1 | i2), Type: token.Number, Pos: op.Pos }

    case token.Xor:
        return token.Token{ Str: fmt.Sprint(i1 ^ i2), Type: token.Number, Pos: op.Pos }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %v\n", op)
        os.Exit(1)
        return token.Token{}
    }
}

func evalUints(op token.Token, i1 uint64, i2 uint64) token.Token {
    switch op.Type {
    case token.Lss:
        if i1 < i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Grt:
        if i1 > i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Leq:
        if i1 <= i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Geq:
        if i1 >= i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Plus:
        return token.Token{ Str: fmt.Sprint(i1 + i2), Type: token.Number, Pos: op.Pos }

    case token.Minus:
        return token.Token{ Str: fmt.Sprint(i1 - i2), Type: token.Number, Pos: op.Pos }

    case token.Mul:
        return token.Token{ Str: fmt.Sprint(i1 * i2), Type: token.Number, Pos: op.Pos }

    case token.Div:
        return token.Token{ Str: fmt.Sprint(i1 / i2), Type: token.Number, Pos: op.Pos }

    case token.Mod:
        return token.Token{ Str: fmt.Sprint(i1 % i2), Type: token.Number, Pos: op.Pos }


    case token.Shl:
        return token.Token{ Str: fmt.Sprint(i1 << i2), Type: token.Number, Pos: op.Pos }

    case token.Shr:
        return token.Token{ Str: fmt.Sprint(i1 >> i2), Type: token.Number, Pos: op.Pos }

    case token.Amp:
        return token.Token{ Str: fmt.Sprint(i1 & i2), Type: token.Number, Pos: op.Pos }

    case token.BitOr:
        return token.Token{ Str: fmt.Sprint(i1 | i2), Type: token.Number, Pos: op.Pos }

    case token.Xor:
        return token.Token{ Str: fmt.Sprint(i1 ^ i2), Type: token.Number, Pos: op.Pos }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %v\n", op)
        os.Exit(1)
        return token.Token{}
    }
}

func toInts(lhs token.Token, rhs token.Token) (rangeErr bool, l int64, r int64) {
    if i,err := strconv.ParseInt(lhs.Str, 0, 64); err == nil {
        l = i
    } else {
        if e,ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
            return true, 0, 0
        }
    }

    if i,err := strconv.ParseInt(rhs.Str, 0, 64); err == nil {
        r = i
    } else {
        if e,ok := err.(*strconv.NumError); ok && e.Err == strconv.ErrRange {
            return true, 0, 0
        }
    }

    return
}

func toUints(lhs token.Token, rhs token.Token) (l uint64, r uint64) {
    if i,err := strconv.ParseUint(lhs.Str, 0, 64); err == nil {
        l = i
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not a valid u64\n", lhs)
        fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
        fmt.Fprintln(os.Stderr, "\t" + lhs.Pos.At())
        os.Exit(1)
    }

    if i,err := strconv.ParseUint(rhs.Str, 0, 64); err == nil {
        r = i
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] %v is not a valid u64\n", rhs)
        fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
        fmt.Fprintln(os.Stderr, "\t" + rhs.Pos.At())
        os.Exit(1)
    }

    return
}

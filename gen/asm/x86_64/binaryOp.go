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

    case token.Lss:
        i1, i2 := toInts(lhs, rhs)

        if i1 < i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Grt:
        i1, i2 := toInts(lhs, rhs)

        if i1 > i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Leq:
        i1, i2 := toInts(lhs, rhs)

        if i1 <= i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Geq:
        i1, i2 := toInts(lhs, rhs)

        if i1 >= i2 {
            return token.Token{ Str: "true", Type: token.Boolean, Pos: op.Pos }
        } else {
            return token.Token{ Str: "false", Type: token.Boolean, Pos: op.Pos }
        }

    case token.Plus:
        i1, i2 := toInts(lhs, rhs)
        return token.Token{ Str: fmt.Sprint(i1 + i2), Type: token.Number, Pos: op.Pos }

    case token.Minus:
        i1, i2 := toInts(lhs, rhs)
        return token.Token{ Str: fmt.Sprint(i1 - i2), Type: token.Number, Pos: op.Pos }

    case token.Mul:
        i1, i2 := toInts(lhs, rhs)
        return token.Token{ Str: fmt.Sprint(i1 * i2), Type: token.Number, Pos: op.Pos }

    case token.Div:
        i1, i2 := toInts(lhs, rhs)
        return token.Token{ Str: fmt.Sprint(i1 / i2), Type: token.Number, Pos: op.Pos }

    case token.Mod:
        i1, i2 := toInts(lhs, rhs)
        return token.Token{ Str: fmt.Sprint(i1 % i2), Type: token.Number, Pos: op.Pos }

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %v\n", op)
        os.Exit(1)
        return token.Token{}
    }
}

func toInts(lhs token.Token, rhs token.Token) (l int, r int) {
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

    l,_ = strconv.Atoi(lhs.Str)
    r,_ = strconv.Atoi(rhs.Str)

    return
}

package arith

import (
    "os"
    "fmt"
    "strconv"
    "gorec/vars"
    "gorec/token"
    "gorec/asm/x86_64"
)

// results in A-register

func BinaryOp(file *os.File, opType token.TokenType, src string, size int) {
    var s string
    switch opType {
    case token.Eql:
        s = asm.Eql(asm.GetReg(asm.RegA, size), src)
    case token.Neq:
        s = asm.Neq(asm.GetReg(asm.RegA, size), src)
    case token.Lss:
        s = asm.Lss(asm.GetReg(asm.RegA, size), src)
    case token.Grt:
        s = asm.Grt(asm.GetReg(asm.RegA, size), src)
    case token.Leq:
        s = asm.Leq(asm.GetReg(asm.RegA, size), src)
    case token.Geq:
        s = asm.Geq(asm.GetReg(asm.RegA, size), src)

    case token.Plus:
        s = asm.Add(src, size)
    case token.Minus:
        s = asm.Sub(src, size)
    case token.Mul:
        s = asm.Mul(src, size)
    case token.Div:
        s = asm.Div(src, size)
    case token.Mod:
        s = asm.Mod(src, size)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %v\n", opType)
        os.Exit(1)
    }

    vars.Write(file, s)
}

func BinaryOpReg(file *os.File, opType token.TokenType, reg asm.RegGroup, size int) {
    BinaryOp(file, opType, asm.GetReg(reg, size), size)
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

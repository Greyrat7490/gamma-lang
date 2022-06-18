package arith

import (
	"os"
	"fmt"
	"gorec/token"
	"gorec/vars"
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
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %s\n", opType.Readable())
        os.Exit(1)
    }

    vars.Write(file, s)
}

func BinaryOpReg(file *os.File, opType token.TokenType, reg asm.RegGroup, size int) {
    BinaryOp(file, opType, asm.GetReg(reg, size), size)
}

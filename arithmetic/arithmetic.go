package arith

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/token"
)

// results in A-register

func BinaryOp(asm *os.File, opType token.TokenType, src string, size int) {
    switch opType {
    case token.Eql:
        eql(asm, vars.GetReg(vars.RegA, size), src)
    case token.Neq:
        neq(asm, vars.GetReg(vars.RegA, size), src)
    case token.Lss:
        lss(asm, vars.GetReg(vars.RegA, size), src)
    case token.Grt:
        grt(asm, vars.GetReg(vars.RegA, size), src)
    case token.Leq:
        leq(asm, vars.GetReg(vars.RegA, size), src)
    case token.Geq:
        geq(asm, vars.GetReg(vars.RegA, size), src)

    case token.Plus:
        add(asm, src, size)
    case token.Minus:
        sub(asm, src, size)
    case token.Mul:
        mul(asm, src, size)
    case token.Div:
        div(asm, src, size)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %s\n", opType.Readable())
        os.Exit(1)
    }
}

func BinaryOpReg(asm *os.File, opType token.TokenType, reg vars.RegGroup, size int) {
    switch opType {
    case token.Eql:
        eql(asm, vars.GetReg(vars.RegA, size), vars.GetReg(reg, size))
    case token.Neq:
        neq(asm, vars.GetReg(vars.RegA, size), vars.GetReg(reg, size))
    case token.Lss:
        lss(asm, vars.GetReg(vars.RegA, size), vars.GetReg(reg, size))
    case token.Grt:
        grt(asm, vars.GetReg(vars.RegA, size), vars.GetReg(reg, size))
    case token.Leq:
        leq(asm, vars.GetReg(vars.RegA, size), vars.GetReg(reg, size))
    case token.Geq:
        geq(asm, vars.GetReg(vars.RegA, size), vars.GetReg(reg, size))

    case token.Plus:
        addReg(asm, reg, size)
    case token.Minus:
        subReg(asm, reg, size)
    case token.Mul:
        mulReg(asm, reg, size)
    case token.Div:
        divReg(asm, reg, size)
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %s\n", opType.Readable())
        os.Exit(1)
    }
}

func eql(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "sete al\n")
}
func neq(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "setne al\n")
}
func lss(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "setl al\n")
}
func grt(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "setg al\n")
}
func leq(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "setle al\n")
}
func geq(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "setge al\n")
}


func add(asm *os.File, src string, size int) {
    vars.Write(asm, fmt.Sprintf("add %s, %s\n", vars.GetReg(vars.RegA, size), src))
}
func addReg(asm *os.File, src vars.RegGroup, size int) {
    add(asm, vars.GetReg(src, size), size)
}

func sub(asm *os.File, src string, size int) {
    vars.Write(asm, fmt.Sprintf("sub %s, %s\n", vars.GetReg(vars.RegA, size), src))
}
func subReg(asm *os.File, src vars.RegGroup, size int) {
    sub(asm, vars.GetReg(src, size), size)
}

func mul(asm *os.File, src string, size int) {
    push(asm, vars.RegB)
    push(asm, vars.RegD)

    vars.Write(asm, fmt.Sprintf("mov %s, %s\n", vars.GetReg(vars.RegB, size), src))
    vars.Write(asm, fmt.Sprintf("imul %s\n", vars.GetReg(vars.RegB, size)))

    pop(asm, vars.RegD)
    pop(asm, vars.RegB)
}
func mulReg(asm *os.File, src vars.RegGroup, size int) {
    mul(asm, vars.GetReg(src, size), size)
}

func div(asm *os.File, src string, size int) {
    push(asm, vars.RegD)
    push(asm, vars.RegB)

    // TODO: check if dest is signed or unsigned (use either idiv or div)
    // for now only signed integers are supported
    vars.Write(asm, fmt.Sprintf("mov %s, %s\n", vars.GetReg(vars.RegB, size), src))
    if size == 8 {
        vars.Write(asm, "cqo\n") // sign extend rax into rdx (div with 64bit regs -> 128bit div)
    } else if size == 4 {
        vars.Write(asm, "cdq\n") // sign extend eax into edx (div with 32bit regs -> 64bit div)
    }
    vars.Write(asm, fmt.Sprintf("idiv %s\n", vars.GetReg(vars.RegB, size)))

    pop(asm, vars.RegB)
    pop(asm, vars.RegD)
}
func divReg(asm *os.File, src vars.RegGroup, size int) {
    div(asm, vars.GetReg(src, size), size)
}

func push(asm *os.File, reg vars.RegGroup) {
    vars.Write(asm, fmt.Sprintf("push %s\n", vars.GetReg(reg, 8)))
}
func pop(asm *os.File, reg vars.RegGroup) {
    vars.Write(asm, fmt.Sprintf("pop %s\n", vars.GetReg(reg, 8)))
}

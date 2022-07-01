package asm

import (
    "fmt"
)

func Eql(lhs string, rhs string) string {
    return fmt.Sprintf("cmp %s, %s\nsete al\n", lhs, rhs)
}
func Neq(lhs string, rhs string) string{
    return fmt.Sprintf("cmp %s, %s\nsetne al\n", lhs, rhs)
}
func Lss(lhs string, rhs string) string {
    return fmt.Sprintf("cmp %s, %s\nsetl al\n", lhs, rhs)
}
func Grt(lhs string, rhs string) string {
    return fmt.Sprintf("cmp %s, %s\nsetg al\n", lhs, rhs)
}
func Leq(lhs string, rhs string) string {
    return fmt.Sprintf("cmp %s, %s\nsetle al\n", lhs, rhs)
}
func Geq(lhs string, rhs string) string {
    return fmt.Sprintf("cmp %s, %s\nsetge al\n", lhs, rhs)
}


func Neg(src string, size int) string {
    return fmt.Sprintf("neg %s\n", GetReg(RegA, size))
}
func Add(src string, size int) string {
    return fmt.Sprintf("add %s, %s\n", GetReg(RegA, size), src)
}
func Sub(src string, size int) string {
    return fmt.Sprintf("sub %s, %s\n", GetReg(RegA, size), src)
}
func Mul(src string, size int) string {
    return Push(RegB) +
    Push(RegD) +

    fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src) +
    fmt.Sprintf("imul %s\n", GetReg(RegB, size)) +

    Pop(RegD) +
    Pop(RegB)
}
func Div(src string, size int) string {
    res := Push(RegD) +
    Push(RegB) +

    // TODO: check if dest is signed or unsigned (use either idiv or div)
    // for now only signed integers are supported
    fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src)
    if size == 8 {
        res += "cqo\n" // sign extend rax into rdx (div with 64bit regs -> 128bit div)
    } else if size == 4 {
        res += "cdq\n" // sign extend eax into edx (div with 32bit regs -> 64bit div)
    }
    res += fmt.Sprintf("idiv %s\n", GetReg(RegB, size)) +

    Pop(RegB) +
    Pop(RegD)

    return res
}
func Mod(src string, size int) string {
    res := Push(RegB) + Push(RegD) +

    fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src)
    if size == 8 {
        res += "cqo\n"
    } else if size == 4 {
        res += "cdq\n"
    }
    res += fmt.Sprintf("idiv %s\n", GetReg(RegB, size)) +

    fmt.Sprintf("mov %s, %s\n", GetReg(RegA, size), GetReg(RegD, size)) +
    Pop(RegB) + Pop(RegD)

    return res
}
func Push(reg RegGroup) string {
    return fmt.Sprintf("push %s\n", GetReg(reg, 8))
}
func Pop(reg RegGroup) string {
    return fmt.Sprintf("pop %s\n", GetReg(reg, 8))
}

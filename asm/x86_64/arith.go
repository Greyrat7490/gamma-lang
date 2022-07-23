package asm

import (
    "os"
    "fmt"
)

func Eql(file *os.File, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsete al\n", lhs, rhs))
}
func Neq(file *os.File, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetne al\n", lhs, rhs))
}
func Lss(file *os.File, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetl al\n", lhs, rhs))
}
func Grt(file *os.File, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetg al\n", lhs, rhs))
}
func Leq(file *os.File, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetle al\n", lhs, rhs))
}
func Geq(file *os.File, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetge al\n", lhs, rhs))
}


func Neg(file *os.File, src string, size int) {
    file.WriteString(fmt.Sprintf("neg %s\n", GetReg(RegA, size)))
}
func Add(file *os.File, src string, size int) {
    file.WriteString(fmt.Sprintf("add %s, %s\n", GetReg(RegA, size), src))
}
func Sub(file *os.File, src string, size int) {
    file.WriteString(fmt.Sprintf("sub %s, %s\n", GetReg(RegA, size), src))
}
func Mul(file *os.File, src string, size int) {
    Push(file, RegB)
    Push(file, RegD)

    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src))
    file.WriteString(fmt.Sprintf("imul %s\n", GetReg(RegB, size)))

    Pop(file, RegD)
    Pop(file, RegB)
}
func Div(file *os.File, src string, size int) {
    Push(file, RegD)
    Push(file, RegB)

    // TODO: check if dest is signed or unsigned (use either idiv or div)
    // for now only signed integers are supported
    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src))
    if size == 8 {
        file.WriteString("cqo\n") // sign extend rax into rdx (div with 64bit regs -> 128bit div)
    } else if size == 4 {
        file.WriteString("cdq\n") // sign extend eax into edx (div with 32bit regs -> 64bit div)
    }
    file.WriteString(fmt.Sprintf("idiv %s\n", GetReg(RegB, size)))

    Pop(file, RegB)
    Pop(file, RegD)
}
func Mod(file *os.File, src string, size int) {
    Push(file, RegB)
    Push(file, RegD)
    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src))

    if size == 8 {
        file.WriteString("cqo\n")
    } else if size == 4 {
        file.WriteString("cdq\n")
    }

    file.WriteString(fmt.Sprintf("idiv %s\n", GetReg(RegB, size)))

    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegA, size), GetReg(RegD, size)))
    Pop(file, RegB)
    Pop(file, RegD)
}
func Push(file *os.File, reg RegGroup) {
    file.WriteString(fmt.Sprintf("push %s\n", GetReg(reg, 8)))
}
func Pop(file *os.File, reg RegGroup) {
    file.WriteString(fmt.Sprintf("pop %s\n", GetReg(reg, 8)))
}

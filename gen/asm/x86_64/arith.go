package asm

import (
    "fmt"
    "bufio"
)

func Eql(file *bufio.Writer, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsete al\n", lhs, rhs))
}
func Neq(file *bufio.Writer, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetne al\n", lhs, rhs))
}
func Lss(file *bufio.Writer, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetl al\n", lhs, rhs))
}
func Grt(file *bufio.Writer, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetg al\n", lhs, rhs))
}
func Leq(file *bufio.Writer, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetle al\n", lhs, rhs))
}
func Geq(file *bufio.Writer, lhs string, rhs string) {
    file.WriteString(fmt.Sprintf("cmp %s, %s\nsetge al\n", lhs, rhs))
}


func Neg(file *bufio.Writer, size uint) {
    file.WriteString(fmt.Sprintf("neg %s\n", GetReg(RegA, size)))
}
func Add(file *bufio.Writer, src string, size uint) {
    file.WriteString(fmt.Sprintf("add %s, %s\n", GetReg(RegA, size), src))
}
func Sub(file *bufio.Writer, src string, size uint) {
    file.WriteString(fmt.Sprintf("sub %s, %s\n", GetReg(RegA, size), src))
}
func Mul(file *bufio.Writer, src string, size uint, signed bool) {
    PushReg(file, RegD)

    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src))
    if signed {
        file.WriteString(fmt.Sprintf("imul %s\n", GetReg(RegB, size)))
    } else {
        file.WriteString(fmt.Sprintf("mul %s\n", GetReg(RegB, size)))
    }

    PopReg(file, RegD)
}
func Div(file *bufio.Writer, src string, size uint, signed bool) {
    PushReg(file, RegD)

    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src))
    if signed {
        if size == 8 {
            file.WriteString("cqo\n") // sign extend rax into rdx (div with 64bit regs -> 128bit div)
        } else if size == 4 {
            file.WriteString("cdq\n") // sign extend eax into edx (div with 32bit regs -> 64bit div)
        }
        file.WriteString(fmt.Sprintf("idiv %s\n", GetReg(RegB, size)))
    } else {
        file.WriteString(fmt.Sprintf("xor %s, %s\n", GetReg(RegD, size), GetReg(RegD, size)))
        file.WriteString(fmt.Sprintf("div %s\n", GetReg(RegB, size)))
    }

    PopReg(file, RegD)
}
func Mod(file *bufio.Writer, src string, size uint, signed bool) {
    PushReg(file, RegD)
    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src))

    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegB, size), src))
    if signed {
        if size == 8 {
            file.WriteString("cqo\n")
        } else if size == 4 {
            file.WriteString("cdq\n")
        }
        file.WriteString(fmt.Sprintf("idiv %s\n", GetReg(RegB, size)))
    } else {
        file.WriteString(fmt.Sprintf("xor %s, %s\n", GetReg(RegD, size), GetReg(RegD, size)))
        file.WriteString(fmt.Sprintf("div %s\n", GetReg(RegB, size)))
    }

    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegA, size), GetReg(RegD, size)))
    PopReg(file, RegD)
}

func Not(file *bufio.Writer, size uint) {
    file.WriteString(fmt.Sprintf("not %s\n", GetReg(RegA, size)))
}
func And(file *bufio.Writer, src string, size uint) {
    file.WriteString(fmt.Sprintf("and %s, %s\n", GetReg(RegA, size), src))
}
func Or(file *bufio.Writer, src string, size uint) {
    file.WriteString(fmt.Sprintf("or %s, %s\n", GetReg(RegA, size), src))
}
func Xor(file *bufio.Writer, src string, size uint) {
    file.WriteString(fmt.Sprintf("xor %s, %s\n", GetReg(RegA, size), src))
}
func Shl(file *bufio.Writer, src string, size uint) {
    file.WriteString(fmt.Sprintf("shl %s, %s\n", GetReg(RegA, size), src))
}
func Shr(file *bufio.Writer, src string, size uint) {
    file.WriteString(fmt.Sprintf("shr %s, %s\n", GetReg(RegA, size), src))
}

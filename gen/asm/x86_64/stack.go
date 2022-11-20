package asm

import (
    "fmt"
    "bufio"
    "gamma/types/addr"
)

func PushVal(file *bufio.Writer, val string) {
    file.WriteString(fmt.Sprintf("push %s\n", val))
}

func PopVal(file *bufio.Writer, val string) {
    file.WriteString(fmt.Sprintf("pop %s\n", val))
}

func PushReg(file *bufio.Writer, reg RegGroup) {
    file.WriteString(fmt.Sprintf("push %s\n", GetReg(reg, 8)))
}

func PopReg(file *bufio.Writer, reg RegGroup) {
    file.WriteString(fmt.Sprintf("pop %s\n", GetReg(reg, 8)))
}

func PushDeref(file *bufio.Writer, addr addr.Addr) {
    file.WriteString(fmt.Sprintf("push QWORD [%s]\n", addr))
}

func PopDeref(file *bufio.Writer, addr addr.Addr) {
    file.WriteString(fmt.Sprintf("pop QWORD [%s]\n", addr))
}

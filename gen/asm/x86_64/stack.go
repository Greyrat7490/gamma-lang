package asm

import (
    "os"
    "fmt"
    "gamma/types/addr"
)

func PushVal(file *os.File, val string) {
    file.WriteString(fmt.Sprintf("push %s\n", val))
}

func PopVal(file *os.File, val string) {
    file.WriteString(fmt.Sprintf("pop %s\n", val))
}

func PushReg(file *os.File, reg RegGroup) {
    file.WriteString(fmt.Sprintf("push %s\n", GetReg(reg, 8)))
}

func PopReg(file *os.File, reg RegGroup) {
    file.WriteString(fmt.Sprintf("pop %s\n", GetReg(reg, 8)))
}

func PushDeref(file *os.File, addr addr.Addr) {
    file.WriteString(fmt.Sprintf("push QWORD [%s]\n", addr))
}

func PopDeref(file *os.File, addr addr.Addr) {
    file.WriteString(fmt.Sprintf("pop QWORD [%s]\n", addr))
}

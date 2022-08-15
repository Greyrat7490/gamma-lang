package asm

import (
    "os"
    "fmt"
)

func Push(file *os.File, reg RegGroup) {
    file.WriteString(fmt.Sprintf("push %s\n", GetReg(reg, 8)))
}

func Pop(file *os.File, reg RegGroup) {
    file.WriteString(fmt.Sprintf("pop %s\n", GetReg(reg, 8)))
}

func PushDeref(file *os.File, addr string) {
    file.WriteString(fmt.Sprintf("push QWORD [%s]\n", addr))
}

func PopDeref(file *os.File, addr string) {
    file.WriteString(fmt.Sprintf("pop QWORD [%s]\n", addr))
}

package asm

import (
    "fmt"
    "bufio"
)

func Lea(file *bufio.Writer, dest RegGroup, addr string, size uint) {
    file.WriteString(fmt.Sprintf("lea %s, [%s]\n", GetReg(dest, size), addr))
}

func LeaOffset(file *bufio.Writer, dest RegGroup, offset int64, size uint) {
    file.WriteString(fmt.Sprintf("lea %s, [%s+%d]\n", GetReg(dest, size), GetReg(dest, size), offset))
}

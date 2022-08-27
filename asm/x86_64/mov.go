package asm

import (
    "os"
    "fmt"
)

func MovRegVal(file *os.File, dest RegGroup, size uint, val string) {
    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(dest, size), val))
}
func MovRegReg(file *os.File, dest RegGroup, src RegGroup, size uint) {
    if GetSize(dest, size) > size {
        file.WriteString(fmt.Sprintf("movzx %s, %s\n", GetReg(dest, size), GetReg(src, size)))
    } else {
        file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(dest, size), GetReg(src, size)))
    }
}
func MovRegDeref(file *os.File, dest RegGroup, addr string, size uint) {
    if GetSize(dest, size) > size {
        file.WriteString(fmt.Sprintf("movzx %s, %s [%s]\n", GetReg(dest, size), GetWord(size), addr))
    } else {
        file.WriteString(fmt.Sprintf("mov %s, %s [%s]\n", GetReg(dest, size), GetWord(size), addr))
    }
}

func MovDerefVal(file *os.File, addr string, size uint, val string) {
    file.WriteString(fmt.Sprintf("mov %s [%s], %s\n", GetWord(size), addr, val))
}
func MovDerefReg(file *os.File, addr string, size uint, reg RegGroup) {
    srcSize := GetSize(reg, size)

    if size < srcSize {
        file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(RegA, srcSize), GetReg(reg, srcSize)))
        file.WriteString(fmt.Sprintf("mov %s [%s], %s\n", GetWord(size), addr, GetReg(RegA, size)))
    } else {
        file.WriteString(fmt.Sprintf("mov %s [%s], %s\n", GetWord(size), addr, GetReg(reg, size)))
    }
}
func MovDerefDeref(file *os.File, dest string, src string, size uint, reg RegGroup) {
    MovRegDeref(file, reg, src, size)
    MovDerefReg(file, dest, size, reg)
}

func DerefRax(file *os.File, size uint) {
    file.WriteString(fmt.Sprintf("mov %s, %s [rax]\n", GetReg(RegA, size), GetWord(size)))
}

func OffsetAddr(baseAddr string, offset int) string {
    if offset == 0 {
        return baseAddr
    }

    if offset > 0 {
        return fmt.Sprintf("%s+%d", baseAddr, offset)
    }

    return fmt.Sprintf("%s%d", baseAddr, offset)
}

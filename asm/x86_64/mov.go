package asm

import (
    "fmt"
)

func MovRegVal(dest RegGroup, size int, val string) string {
    return fmt.Sprintf("mov %s, %s\n", GetReg(dest, size), val)
}
func MovRegReg(dest RegGroup, src RegGroup, size int) string {
    return fmt.Sprintf("mov %s, %s\n", GetReg(dest, size), GetReg(src, size))
}
func MovRegDeref(dest RegGroup, addr string, size int) string {
    return fmt.Sprintf("mov %s, %s [%s]\n", GetReg(dest, size), GetWord(size), addr)
}

func MovDerefVal(addr string, size int, val string) string {
    return fmt.Sprintf("mov %s [%s], %s\n", GetWord(size), addr, val)
}
func MovDerefReg(addr string, size int, reg RegGroup) string {
    return fmt.Sprintf("mov %s [%s], %s\n", GetWord(size), addr, GetReg(reg, size))
}
func MovDerefDeref(dest string, src string, size int, reg RegGroup) string {
    return MovRegDeref(reg, src, size) + MovDerefReg(dest, size, reg)
}

func MovLocVarVal(offset int, size int, val string) string {
    return fmt.Sprintf("mov %s [rbp-%d], %s\n", GetWord(size), offset, val)
}
func MovLocVarReg(offset int, size int, reg RegGroup) string {
    return fmt.Sprintf("mov %s [rbp-%d], %s\n", GetWord(size), offset, GetReg(reg, size))
}

func DerefRax(size int) string {
    return fmt.Sprintf("mov %s, %s [rax]\n", GetReg(RegA, size), GetWord(size))
}

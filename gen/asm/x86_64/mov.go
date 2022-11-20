package asm

import (
    "fmt"
    "bufio"
    "gamma/types"
    "gamma/types/addr"
)

func MovRegVal(file *bufio.Writer, dest RegGroup, size uint, val string) {
    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(dest, size), val))
}

func MovRegReg(file *bufio.Writer, dest RegGroup, src RegGroup, size uint) {
    file.WriteString(fmt.Sprintf("mov %s, %s\n", GetReg(dest, size), GetReg(src, size)))
}

func MovRegDeref(file *bufio.Writer, dest RegGroup, addr addr.Addr, size uint, signed bool) {
    file.WriteString(fmt.Sprintf("mov%s %s, %s [%s]\n", extend32(size, signed), GetReg(dest, size), GetWord(size), addr))
}

func MovDerefVal(file *bufio.Writer, addr addr.Addr, size uint, val string) {
    file.WriteString(fmt.Sprintf("mov %s [%s], %s\n", GetWord(size), addr, val))
}

func MovDerefReg(file *bufio.Writer, addr addr.Addr, size uint, reg RegGroup) {
    file.WriteString(fmt.Sprintf("mov %s [%s], %s\n", GetWord(size), addr, GetAnyReg(reg, size)))
}

func MovDerefDeref(file *bufio.Writer, dest addr.Addr, src addr.Addr, size uint, reg RegGroup, signed bool) {
    MovRegDeref(file, reg, src, size, signed)
    MovDerefReg(file, dest, size, reg)
}

func DerefRax(file *bufio.Writer, size uint, signed bool) {
    file.WriteString(fmt.Sprintf("mov%s %s, %s [rax]\n", extend32(size, signed), GetReg(RegA, size), GetWord(size)))
}


func MovRegRegExtend(file *bufio.Writer, dest RegGroup, destSize uint, src RegGroup, srcSize uint, signed bool) {
    var ext string
    destSize, ext = extend(destSize, srcSize, signed)
    file.WriteString(fmt.Sprintf("mov%s %s, %s\n", ext, GetReg(dest, destSize), GetAnyReg(src, srcSize)))
}

func MovRegDerefExtend(file *bufio.Writer, dest RegGroup, destSize uint, addr addr.Addr, srcSize uint, signed bool) {
    var ext string
    destSize, ext = extend(destSize, srcSize, signed)
    file.WriteString(fmt.Sprintf("mov%s %s, %s [%s]\n", ext, GetReg(dest, destSize), GetWord(srcSize), addr))
}


func extend32(size uint, signed bool) string {
    if size < types.I32_Size {
        if signed {
            return "sx"
        } else {
            return "zx"
        }
    }

    return ""
}

func extend(destSize uint, srcSize uint, signed bool) (uint, string) {
    if destSize > srcSize {
        if !signed {
            if destSize == types.Ptr_Size {
                return types.I32_Size, extend32(srcSize, signed)
            }

            return destSize, "zx"
        } else {
            return destSize, "sx"
        }
    } else {
        return destSize, extend32(srcSize, signed)
    }
}

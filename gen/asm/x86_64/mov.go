package asm

import (
    "os"
    "fmt"
    "gamma/types"
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


func MovRegRegExtend(file *os.File, dest RegGroup, destSize uint, src RegGroup, srcSize uint, signed bool) {
    destSize = GetSize(dest, destSize)
    srcSize = GetSize(src, srcSize)

    var mov string
    destSize, srcSize, mov = extend(destSize, srcSize, signed)
    file.WriteString(fmt.Sprintf("%s %s, %s\n", mov, GetReg(dest, destSize), GetReg(src, srcSize)))
}

func MovRegDerefExtend(file *os.File, dest RegGroup, destSize uint, addr string, srcSize uint, signed bool) {
    destSize = GetSize(dest, destSize)

    var mov string
    destSize, srcSize, mov = extend(destSize, srcSize, signed)
    file.WriteString(fmt.Sprintf("%s %s, %s [%s]\n", mov, GetReg(dest, destSize), GetWord(srcSize), addr))
}

func extend(destSize uint, srcSize uint, signed bool) (dstSz uint, srcSz uint, mov string) {
    mov = "mov"

    if destSize > srcSize {
        if !signed {
            if destSize == types.Ptr_Size {
                destSize = types.I32_Size
                if types.I32_Size <= srcSize {
                    return destSize, srcSize, mov
                }
            }

            mov += "zx"
        } else {
            mov += "sx"
        }

    } else if destSize < srcSize {
        srcSize = destSize
    }

    return destSize, srcSize, mov
}

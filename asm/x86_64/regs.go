package asm

import "fmt"

var regs [][]string = [][]string{
    { "al", "ax", "eax", "rax" },
    { "dl", "dx", "edx", "rdx" },
    { "bl", "bx", "ebx", "rbx" },
    { "cl", "cx", "ecx", "rcx" },

    { "di", "edi", "rdi" },
    { "si", "esi", "rsi" },

    { "r8d",  "r8" },
    { "r9d",  "r9" },
    { "r10d", "r10" },
    { "r11d", "r11" },
}

var words     []string = []string{ "BYTE", "WORD", "DWORD", "QWORD" }
var dataSizes []string = []string{ "db", "dw", "dd", "dq" }
var bssSizes  []string = []string{ "resb", "resw", "resd", "resq" }

type RegGroup = uint8
const (
    RegA   RegGroup = iota
    RegD   RegGroup = iota
    RegB   RegGroup = iota
    RegC   RegGroup = iota

    RegDi  RegGroup = iota
    RegSi  RegGroup = iota

    RegR8  RegGroup = iota
    RegR9  RegGroup = iota
    RegR10 RegGroup = iota
    RegR11 RegGroup = iota
)

func GetWord(bytes uint) string {
    if bytes == 8 {
        return words[3]
    }
    return words[bytes/2]
}
func GetDataSize(bytes uint) string {
    if bytes == 8 {
        return dataSizes[3]
    }
    return dataSizes[bytes/2]
}
func GetBssSize(bytes uint) string {
    if bytes == 8 {
        return bssSizes[3]
    }
    return bssSizes[bytes/2]
}

func GetReg(g RegGroup, size uint) string {
    if g >= RegR8 {
        if size == 8 {
            return regs[g][1]
        }
        return regs[g][0]
    }

    if g >= RegDi {
        if size == 8 {
            return regs[g][2]
        }
        if size == 1 {
            return regs[g][0]
        }
        return regs[g][size / 2 - 1]
    }

    if size == 8 {
        return regs[g][3]
    }
    return regs[g][size / 2]
}

func GetOffsetedReg(g RegGroup, size uint, offset int) string {
    reg := GetReg(g, size)

    if offset == 0 {
        return reg
    }

    if offset > 0 {
        return fmt.Sprintf("%s+%d", reg, offset)
    }

    return fmt.Sprintf("%s%d", reg, offset)
}

func GetSize(g RegGroup, size uint) uint {
    switch {
    case g >= RegR8 && size < 8:
        return 4
    case g >= RegDi && size < 2:
        return 2
    default:
        return size
    }
}

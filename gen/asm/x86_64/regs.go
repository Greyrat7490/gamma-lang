package asm

import (
    "fmt"
    "bufio"
    "gamma/types/addr"
)

var regs [][]string = [][]string{
    { "al", "ax", "eax", "rax" },
    { "dl", "dx", "edx", "rdx" },
    { "bl", "bx", "ebx", "rbx" },
    { "cl", "cx", "ecx", "rcx" },

    { "dil", "di", "edi", "rdi" },
    { "sil", "si", "esi", "rsi" },

    { "r8b",  "r8w",  "r8d",  "r8" },
    { "r9b",  "r9w",  "r9d",  "r9" },
    { "r10b", "r10w", "r10d", "r10" },
    { "r11b", "r11w", "r11d", "r11" },

    { "spl", "sp", "esp", "rsp" },
    { "bpl", "bp", "ebp", "rbp" },
}

var used []bool = make([]bool, RegCount)
var preserve []bool = make([]bool, RegCount)

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

    RegSp RegGroup = iota
    RegBp RegGroup = iota

    RegCount RegGroup = iota
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
    if size == 8 {
        return regs[g][3]
    }
    return regs[g][2]
}

func GetAnyReg(g RegGroup, size uint) string {
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

func RegAsAddr(reg RegGroup) addr.Addr {
    return addr.Addr{ BaseAddr: GetReg(reg, 8) }
}

func IsUsed(reg RegGroup) bool {
    return used[reg]
}

func UseReg(reg RegGroup) {
    used[reg] = true
}

func FreeReg(reg RegGroup) {
    used[reg] = false
}

func SaveReg(file *bufio.Writer, reg RegGroup) {
    if used[reg] {
        preserve[reg] = true
        PushReg(file, reg)
        used[reg] = false
    }
}

func RestoreReg(file *bufio.Writer, reg RegGroup) {
    if preserve[reg] {
        preserve[reg] = false
        PopReg(file, reg)
        used[reg] = true
    }
}

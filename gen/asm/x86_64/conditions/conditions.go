package cond

import (
    "os"
    "fmt"
    "gamma/types/addr"
)

var ifCount uint = 0

func ResetCount() {
    ifCount = 0
    logCount = 0
    caseCount = 0
    switchCount = 0
}

func IfVar(file *os.File, addr addr.Addr, hasElse bool) uint {
    ifCount++

    file.WriteString(fmt.Sprintf("cmp BYTE [%s], 1\n", addr))
    if hasElse {
        file.WriteString(fmt.Sprintf("jne .else%d\n", ifCount))
    } else {
        file.WriteString(fmt.Sprintf("jne .if%dEnd\n", ifCount))
    }

    return ifCount
}

func IfExpr(file *os.File, hasElse bool) uint {
    ifCount++
    file.WriteString("cmp al, 1\n")
    if hasElse {
        file.WriteString(fmt.Sprintf("jne .else%d\n", ifCount))
    } else {
        file.WriteString(fmt.Sprintf("jne .if%dEnd\n", ifCount))
    }
    return ifCount
}

func IfEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf(".if%dEnd:\n", count))
}

func ElseStart(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf("jmp .else%dEnd\n", count))
    file.WriteString(fmt.Sprintf(".else%d:\n", count))
}

func ElseEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf(".else%dEnd:\n", count))
}

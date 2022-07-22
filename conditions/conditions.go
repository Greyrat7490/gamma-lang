package cond

import (
    "os"
    "fmt"
    "gorec/vars"
)

var ifCount uint = 0

func ResetCount() {
    ifCount = 0
    logCount = 0
    caseCount = 0
    switchCount = 0
}

func IfVar(file *os.File, v vars.Var, hasElse bool) uint {
    ifCount++

    file.WriteString(fmt.Sprintf("cmp BYTE [%s], 1\n", v.Addr(0)))
    if hasElse {
        file.WriteString("pushfq\n")
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
        file.WriteString("pushfq\n")
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
    file.WriteString(fmt.Sprintf(".else%d:\n", count))
    file.WriteString("popfq\n")
    file.WriteString(fmt.Sprintf("je .else%dEnd\n", count))
}

func ElseEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf(".else%dEnd:\n", count))
}

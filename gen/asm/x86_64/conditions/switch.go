package cond

import (
    "os"
    "fmt"
    "bufio"
    "gamma/token"
    "gamma/types/addr"
)

var inSwitch bool = false
var inLastCase bool = false
var caseCount uint = 0
var switchCount uint = 0

func InSwitch() bool {
    return inSwitch
}

func StartSwitch() uint {
    inSwitch = true
    switchCount++

    return switchCount
}

func EndSwitch(file *bufio.Writer) {
    inSwitch = false
    inLastCase = false
    file.WriteString(fmt.Sprintf(".switch%dEnd:\n", switchCount))
}

func InLastCase() {
    inLastCase = true
}

func CaseStart(file *bufio.Writer) {
    caseCount++
    file.WriteString(fmt.Sprintf(".case%d:\n", caseCount))
}

func CaseVar(file *bufio.Writer, addr addr.Addr) {
    file.WriteString(fmt.Sprintf("cmp BYTE [%s], 1\n", addr))
    if !inLastCase {
        file.WriteString(fmt.Sprintf("jne .case%d\n", caseCount+1))
    } else {
        file.WriteString(fmt.Sprintf("jne .switch%dEnd\n", switchCount))
    }
}

func CaseExpr(file *bufio.Writer) {
    file.WriteString("cmp al, 1\n")
    if !inLastCase {
        file.WriteString(fmt.Sprintf("jne .case%d\n", caseCount+1))
    } else {
        file.WriteString(fmt.Sprintf("jne .switch%dEnd\n", switchCount))
    }
}

func CaseBody(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf(".case%dBody:\n", caseCount))
}

func CaseBodyEnd(file *bufio.Writer, count uint) {
    file.WriteString(fmt.Sprintf("jmp .switch%dEnd\n", count))
}

func Break(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf("jmp .switch%dEnd\n", switchCount))
}

func Through(file *bufio.Writer, pos token.Pos) {
    if !inSwitch {
        fmt.Fprintln(os.Stderr, "[ERROR] through can only be used inside a switch")
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    if inLastCase {
        fmt.Fprintln(os.Stderr, "[ERROR] through cannot be used in the last case")
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    file.WriteString(fmt.Sprintf("jmp .case%dBody\n", caseCount+1))
}

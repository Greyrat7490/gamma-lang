package cond

import (
    "os"
    "fmt"
    "bufio"
    "gamma/token"
    "gamma/types/addr"
)

type switch_ struct {
    parent *switch_
    count uint
    caseCount uint
    inLastCase bool
}

var switchCount uint = 0
var curSwitch *switch_ = nil

func InSwitch() bool {
    return curSwitch != nil
}

func StartSwitch() {
    switchCount++
    curSwitch = &switch_{ parent: curSwitch, count: switchCount, caseCount: 0, inLastCase: false }
}

func EndSwitch(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf(".switch%dEnd:\n", curSwitch.count))
    curSwitch = curSwitch.parent
}

func InLastCase() {
    curSwitch.inLastCase = true
}

func CaseStart(file *bufio.Writer) {
    curSwitch.caseCount++
    file.WriteString(fmt.Sprintf(".case%d%d:\n", curSwitch.count, curSwitch.caseCount))
}

func CaseVar(file *bufio.Writer, addr addr.Addr) {
    file.WriteString(fmt.Sprintf("cmp BYTE [%s], 1\n", addr))
    if !curSwitch.inLastCase {
        file.WriteString(fmt.Sprintf("jne .case%d%d\n", curSwitch.count, curSwitch.caseCount+1))
    } else {
        file.WriteString(fmt.Sprintf("jne .switch%dEnd\n", curSwitch.count))
    }
}

func CaseExpr(file *bufio.Writer) {
    file.WriteString("cmp al, 1\n")
    if !curSwitch.inLastCase {
        file.WriteString(fmt.Sprintf("jne .case%d%d\n", curSwitch.count, curSwitch.caseCount+1))
    } else {
        file.WriteString(fmt.Sprintf("jne .switch%dEnd\n", curSwitch.count))
    }
}

func CaseBody(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf(".case%d%dBody:\n", curSwitch.count, curSwitch.caseCount))
}

func CaseBodyEnd(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf("jmp .switch%dEnd\n", curSwitch.count))
}

func Break(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf("jmp .switch%dEnd\n", curSwitch.count))
}

func Through(file *bufio.Writer, pos token.Pos) {
    if curSwitch == nil {
        fmt.Fprintln(os.Stderr, "[ERROR] through can only be used inside a switch")
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    if curSwitch.inLastCase {
        fmt.Fprintln(os.Stderr, "[ERROR] through cannot be used in the last case")
        fmt.Fprintln(os.Stderr, "\t" + pos.At())
        os.Exit(1)
    }

    file.WriteString(fmt.Sprintf("jmp .case%d%dBody\n", curSwitch.count , curSwitch.caseCount+1))
}

package cond

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/token"
)

var inSwitch bool = false
var inLastCase bool = false
var caseCount uint = 0
var switchCount uint = 0

func StartSwitch() uint {
    inSwitch = true
    switchCount++

    return switchCount
}

func EndSwitch(file *os.File) {
    inSwitch = false
    inLastCase = false
    file.WriteString(fmt.Sprintf(".switch%dEnd:\n", switchCount))
}

func InLastCase() {
    inLastCase = true
}

func CaseStart(file *os.File) uint {
    caseCount++
    file.WriteString(fmt.Sprintf(".case%d:\n", caseCount))
    return caseCount
}

func CaseIdent(file *os.File, ident token.Token) {
    v := vars.GetVar(ident.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", ident.Str)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    file.WriteString(fmt.Sprintf("cmp BYTE [%s], 1\n", v.Addr(0)))
    if !inLastCase {
        file.WriteString(fmt.Sprintf("jne .case%d\n", caseCount+1))
    }
}

func CaseExpr(file *os.File) {
    file.WriteString("cmp al, 1\n")
    if !inLastCase {
        file.WriteString(fmt.Sprintf("jne .case%d\n", caseCount+1))
    }
}

func Default(file *os.File) {
    caseCount++

    file.WriteString(fmt.Sprintf(".case%d:\n", caseCount))
}

func CaseBody(file *os.File) {
    file.WriteString(fmt.Sprintf(".case%dBody:\n", caseCount))
}

func CaseBodyEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf("jmp .switch%dEnd\n", count))
}

func Through(file *os.File, pos token.Pos) {
    if inSwitch == false {
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

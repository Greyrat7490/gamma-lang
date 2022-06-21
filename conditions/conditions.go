package cond

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
)

var count uint = 0

func ResetCount() {
    count = 0
    logCount = 0
}

func IfIdent(file *os.File, ident token.Token, hasElse bool) uint {
    v := vars.GetVar(ident.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", ident.Str)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    if v.GetType().GetKind() != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"%s\" to be of type bool but got \"%v\"\n", ident.Str, v.GetType())
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    count++

    file.WriteString(fmt.Sprintf("cmp BYTE [%s], 1\n", v.Addr(0)))
    if hasElse {
        file.WriteString("pushfq\n")
        file.WriteString(fmt.Sprintf("jne .else%d\n", count))
    } else {
        file.WriteString(fmt.Sprintf("jne .if%dEnd\n", count))
    }

    return count
}

func IfExpr(file *os.File, hasElse bool) uint {
    count++
    file.WriteString("cmp al, 1\n")
    if hasElse {
        file.WriteString("pushfq\n")
        file.WriteString(fmt.Sprintf("jne .else%d\n", count))
    } else {
        file.WriteString(fmt.Sprintf("jne .if%dEnd\n", count))
    }
    return count
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

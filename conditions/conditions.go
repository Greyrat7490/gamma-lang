package cond

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
)

var count uint = 0

func ResetCount() { count = 0 }

func IfIdent(file *os.File, ident token.Token) uint {
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
    file.WriteString(fmt.Sprintf("jne .if%dEnd\n", count)) // skip block if false

    return count
}

func IfExpr(file *os.File) uint {
    count++
    file.WriteString(fmt.Sprintf("cmp al, 1\njne .if%dEnd\n", count))
    return count
}

func IfEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf(".if%dEnd:\n", count))
}

func IfElseIdent(file *os.File, ident token.Token) uint {
    count := IfIdent(file, ident)
    file.WriteString("pushfq\n")
    return count
}

func IfElseExpr(file *os.File) uint {
    count := IfExpr(file)
    file.WriteString("pushfq\n")
    return count
}

func ElseStart(file *os.File, count uint) {
    file.WriteString("popfq\n")
    file.WriteString(fmt.Sprintf("je .else%dEnd\n", count))
    file.WriteString(fmt.Sprintf(".if%dEnd:\n", count))
}

func IfElseEnd(file *os.File, count uint) {
    file.WriteString(fmt.Sprintf(".else%dEnd:\n", count))
}

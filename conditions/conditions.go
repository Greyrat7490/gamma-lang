package cond

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/vars"
)

var count uint = 0

func ResetCount() { count = 0 }

func IfIdent(asm *os.File, ident token.Token) uint {
    v := vars.GetVar(ident.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", ident.Str)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    if v.Type != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"%s\" to be of type bool but got \"%s\"\n", ident.Str, v.Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    count++

    asm.WriteString(fmt.Sprintf("cmp %s, 1\n", v.Get()))
    asm.WriteString(fmt.Sprintf("jne .if%dEnd\n", count)) // skip block if false

    return count
}

func IfReg(asm *os.File, reg string) uint {
    count++

    asm.WriteString(fmt.Sprintf("cmp %s, 1\n", reg))
    asm.WriteString(fmt.Sprintf("jne .if%dEnd\n", count))

    return count
}

func IfEnd(asm *os.File, count uint) {
    asm.WriteString(fmt.Sprintf(".if%dEnd:\n", count))
}

func IfElseIdent(asm *os.File, ident token.Token) uint {
    count := IfIdent(asm, ident)
    asm.WriteString("pushfq\n")
    return count
}

func IfElseReg(asm *os.File, reg string) uint {
    count := IfReg(asm, reg)
    asm.WriteString("pushfq\n")
    return count
}

func ElseStart(asm *os.File, count uint) {
    asm.WriteString("popfq\n")
    asm.WriteString(fmt.Sprintf("je .else%dEnd\n", count))
    asm.WriteString(fmt.Sprintf(".if%dEnd:\n", count))
}

func IfElseEnd(asm *os.File, count uint) {
    asm.WriteString(fmt.Sprintf(".else%dEnd:\n", count))
}

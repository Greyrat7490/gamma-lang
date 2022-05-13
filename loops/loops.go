package loops

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
)


var whileCount uint = 0
var forCount   uint = 0

func ResetCount() {
    whileCount = 0
    forCount   = 0
}

func WhileStart(asm *os.File) uint {
    whileCount++
    asm.WriteString(fmt.Sprintf(".while%d:\n", whileCount))
    return whileCount
}

func WhileIdent(asm *os.File, ident token.Token) {
    v := vars.GetVar(ident.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", ident.Str)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    if v.Vartype != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"%s\" to be of type bool but got \"%s\"\n", ident.Str, v.Vartype.Readable())
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    asm.WriteString(fmt.Sprintf("cmp %s, 1\n", vars.Registers[v.Regs[0]].Name))
    asm.WriteString(fmt.Sprintf("jne .while%dEnd\n", whileCount))
}

func WhileReg(asm *os.File, reg string) {
    asm.WriteString(fmt.Sprintf("cmp %s, 1\n", reg))
    asm.WriteString(fmt.Sprintf("jne .while%dEnd\n", whileCount))
}

func WhileEnd(asm *os.File, count uint) {
    asm.WriteString(fmt.Sprintf("jmp .while%d\n", count))
    asm.WriteString(fmt.Sprintf(".while%dEnd:\n", count))
}


func ForStart(asm *os.File) uint {
    forCount++
    asm.WriteString(fmt.Sprintf(".for%d:\n", forCount))
    return forCount
}

func ForIdent(asm *os.File, ident token.Token) {
    v := vars.GetVar(ident.Str)

    if v == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", ident.Str)
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    if v.Vartype != types.Bool {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"%s\" to be of type bool but got \"%s\"\n", ident.Str, v.Vartype.Readable())
        fmt.Fprintln(os.Stderr, "\t" + ident.At())
        os.Exit(1)
    }

    asm.WriteString(fmt.Sprintf("cmp %s, 1\n", vars.Registers[v.Regs[0]].Name))
    asm.WriteString(fmt.Sprintf("jne .for%dEnd\n", forCount))
}

func ForReg(asm *os.File, reg string) {
    asm.WriteString(fmt.Sprintf("cmp %s, 1\n", reg))
    asm.WriteString(fmt.Sprintf("jne .for%dEnd\n", forCount))
}

func ForEnd(asm *os.File, count uint) {
    asm.WriteString(fmt.Sprintf("jmp .for%d\n", count))
    asm.WriteString(fmt.Sprintf(".for%dEnd:\n", count))
}

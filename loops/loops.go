package loops

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
)


var count uint = 0

func ResetCount() { count = 0 }

func WhileStart(asm *os.File) uint {
    count++
    asm.WriteString(fmt.Sprintf(".while%d:\n", count))   
    return count
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
    asm.WriteString(fmt.Sprintf("jne .while%dEnd\n", count))
}

func WhileReg(asm *os.File, reg string) {
    asm.WriteString(fmt.Sprintf("cmp %s, 1\n", reg))
    asm.WriteString(fmt.Sprintf("jne .while%dEnd\n", count))
}

func WhileEnd(asm *os.File, count uint) {
    asm.WriteString(fmt.Sprintf("jmp .while%d\n", count))
    asm.WriteString(fmt.Sprintf(".while%dEnd:\n", count))
}

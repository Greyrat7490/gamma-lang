package cond

import (
	"fmt"
	"gorec/token"
	"gorec/types"
	"gorec/vars"
	"os"
)

var ifCount = 0

func ResetCount() {
    ifCount = 0
}

func IfIdent(asm *os.File, ident token.Token) {
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
    asm.WriteString(fmt.Sprintf("jne .if%dEnd\n", ifCount)) // skip block if false
}

func IfReg(asm *os.File, reg string) {
    asm.WriteString(fmt.Sprintf("cmp %s, 1\n", reg))
    asm.WriteString(fmt.Sprintf("jne .if%dEnd\n", ifCount))
}

func IfEnd(asm *os.File) {
    asm.WriteString(fmt.Sprintf(".if%dEnd:\n", ifCount))
    ifCount++
}

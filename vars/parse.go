package vars

import (
    "fmt"
    "gorec/types"
    "gorec/parser"
    "os"
    "strconv"
)

func IsLit(w string) bool {
    if w[0] == '"' && w[len(w) - 1] == '"' {
        return true
    }

    if _, err := strconv.Atoi(w); err == nil {
        return true
    }

    return false
}

func ParseDeclare(words []prs.Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] neither name nor type provided for the variable declaration")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    if len(words) < idx + 2 {
        fmt.Fprintln(os.Stderr, "[ERROR] no name or type provided for the variable")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    // maybe implement shadowing later (TODO)
    if Get(words[idx+1].Str) != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", words[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    t := types.ToType(words[idx+2].Str)
    if t == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", words[idx+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].At())
        os.Exit(1)
    }

    op := prs.Op{ Type: prs.OP_DEC_VAR, Token: words[idx], Operants: []string{ words[idx+1].Str, words[idx+2].Str } }
    prs.Ops = append(prs.Ops, op)

    return idx + 2
}

func ParseDefine(words []prs.Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    value := words[idx+1].Str
    v := words[idx-2].Str

    op := prs.Op{ Type: prs.OP_DEF_VAR, Token: words[idx], Operants: []string{ v, value } }
    prs.Ops = append(prs.Ops, op)

    return idx + 1
}

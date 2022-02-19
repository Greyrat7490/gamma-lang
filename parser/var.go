package prs

import (
    "fmt"
    "gorec/types"
    "os"
)

func prsDecVar(words []Token, idx int) int {
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

    t := types.ToType(words[idx+2].Str)
    if t == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", words[idx+2].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_DEC_VAR, Token: words[idx+1], Operants: []string{ words[idx+1].Str, words[idx+2].Str } }
    Ops = append(Ops, op)

    return idx + 2
}

func prsDefVar(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    value := words[idx+1].Str
    v := words[idx-2].Str
    t := words[idx-2]

    // process sign
    if value == "+" || value == "-" {
        if IsLit(words[idx+2].Str) {
            value += words[idx+2].Str
        } else {
            if value == "+" {
                value = words[idx+2].Str
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] negating a variable is not yet supported\n")
                os.Exit(1)
            }
        }
        idx++
    }

    op := Op{ Type: OP_DEF_VAR, Token: t, Operants: []string{ v, value } }
    Ops = append(Ops, op)

    return idx + 1
}

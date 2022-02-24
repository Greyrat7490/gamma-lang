package prs

import (
    "fmt"
    "os"
)

func prsAdd(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '+' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    destOpIdx := len(Ops)-1
    for isBinaryOp(Ops[destOpIdx].Type) { destOpIdx-- }

    if Ops[destOpIdx].Type == OP_DEF_VAR || Ops[destOpIdx].Type == OP_ASSIGN_VAR {
        op := Op{ Type: OP_ADD, Token: words[idx], Operants: []string{ Ops[destOpIdx].Operants[0], words[idx+1].Str } }
        Ops = append(Ops, op)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    return idx + 1
}

func prsSub(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '-' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    destOpIdx := len(Ops)-1
    for isBinaryOp(Ops[destOpIdx].Type) { destOpIdx-- }

    if Ops[destOpIdx].Type == OP_DEF_VAR || Ops[destOpIdx].Type == OP_ASSIGN_VAR {
        op := Op{ Type: OP_SUB, Token: words[idx], Operants: []string{ Ops[destOpIdx].Operants[0], words[idx+1].Str } }
        Ops = append(Ops, op)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    return idx + 1
}

func prsMul(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '*' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    destOpIdx := len(Ops)-1
    for isBinaryOp(Ops[destOpIdx].Type) { destOpIdx-- }

    if Ops[destOpIdx].Type == OP_DEF_VAR || Ops[destOpIdx].Type == OP_ASSIGN_VAR {
        op := Op{ Type: OP_MUL, Token: words[idx], Operants: []string{ Ops[destOpIdx].Operants[0], words[idx+1].Str } }
        Ops = append(Ops, op)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    return idx + 1
}

func prsDiv(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '/' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    destOpIdx := len(Ops)-1
    for isBinaryOp(Ops[destOpIdx].Type) { destOpIdx-- }

    if Ops[destOpIdx].Type == OP_DEF_VAR || Ops[destOpIdx].Type == OP_ASSIGN_VAR {
        op := Op{ Type: OP_DIV, Token: words[idx], Operants: []string{ Ops[destOpIdx].Operants[0], words[idx+1].Str } }
        Ops = append(Ops, op)
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    return idx + 1
}

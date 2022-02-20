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

    prevOp := Ops[len(Ops)-1]
    dest := prevOp.Operants[0]

    if prevOp.Type != OP_DEF_VAR && prevOp.Type != OP_ASSIGN_VAR{
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_ADD, Token: words[idx], Operants: []string{ dest, words[idx+1].Str } }
    Ops = append(Ops, op)

    return idx + 1
}

func prsSub(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '-' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    prevOp := Ops[len(Ops)-1]
    dest := prevOp.Operants[0]

    if prevOp.Type != OP_DEF_VAR && prevOp.Type != OP_ASSIGN_VAR{
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_SUB, Token: words[idx], Operants: []string{ dest, words[idx+1].Str } }
    Ops = append(Ops, op)

    return idx + 1
}

func prsMul(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '*' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    prevOp := Ops[len(Ops)-1]
    dest := prevOp.Operants[0]

    if prevOp.Type != OP_DEF_VAR && prevOp.Type != OP_ASSIGN_VAR{
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_MUL, Token: words[idx], Operants: []string{ dest, words[idx+1].Str } }
    Ops = append(Ops, op)

    return idx + 1
}

func prsDiv(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '/' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    prevOp := Ops[len(Ops)-1]
    dest := prevOp.Operants[0]

    if prevOp.Type != OP_DEF_VAR && prevOp.Type != OP_ASSIGN_VAR{
        fmt.Fprintln(os.Stderr, "[ERROR] not using result (assigning or defining a var)")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_DIV, Token: words[idx], Operants: []string{ dest, words[idx+1].Str } }
    Ops = append(Ops, op)

    return idx + 1
}

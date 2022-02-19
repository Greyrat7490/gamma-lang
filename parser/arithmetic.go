package prs

import (
    "fmt"
    "os"
    "gorec/types"
)

func prsAdd(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '+' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    t1 := types.TypeOfVal(words[idx-1].Str)
    t2 := types.TypeOfVal(words[idx+1].Str)
    if t1 != types.I32 {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only add 2 i32 values(\"%s\" is \"%s\" not i32)\n", words[idx-1].Str, t1.Readable())
        fmt.Fprintln(os.Stderr, "\t" + words[idx-1].At())
        os.Exit(1)
    }
    if t2 != types.I32 {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only add 2 i32 values(\"%s\" is \"%s\" not i32)\n", words[idx+1].Str, t2.Readable())
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_ADD, Token: words[idx], Operants: []string{ words[idx-4].Str, words[idx+1].Str } }
    Ops = append(Ops, op)

    return idx + 1
}

func prsSub(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '-' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    t1 := types.TypeOfVal(words[idx-1].Str)
    t2 := types.TypeOfVal(words[idx+1].Str)
    if t1 != types.I32 {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only sub 2 i32 values(\"%s\" is \"%s\" not i32)\n", words[idx-1].Str, t1.Readable())
        fmt.Fprintln(os.Stderr, "\t" + words[idx-1].At())
        os.Exit(1)
    }
    if t2 != types.I32 {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only sub 2 i32 values(\"%s\" is \"%s\" not i32)\n", words[idx+1].Str, t2.Readable())
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_SUB, Token: words[idx], Operants: []string{ words[idx-4].Str, words[idx+1].Str } }
    Ops = append(Ops, op)

    return idx + 1
}

func prsMul(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '*' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    t1 := types.TypeOfVal(words[idx-1].Str)
    t2 := types.TypeOfVal(words[idx+1].Str)
    if t1 != types.I32 {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only mul 2 i32 values(\"%s\" is \"%s\" not i32)\n", words[idx-1].Str, t1.Readable())
        fmt.Fprintln(os.Stderr, "\t" + words[idx-1].At())
        os.Exit(1)
    }
    if t2 != types.I32 {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only mul 2 i32 values(\"%s\" is \"%s\" not i32)\n", words[idx+1].Str, t2.Readable())
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_MUL, Token: words[idx], Operants: []string{ words[idx-4].Str, words[idx+1].Str } }
    Ops = append(Ops, op)

    return idx + 1
}

func prsDiv(words []Token, idx int) int {
    if len(words) < idx + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] '/' needs 2 operants")
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    t1 := types.TypeOfVal(words[idx-1].Str)
    t2 := types.TypeOfVal(words[idx+1].Str)
    if t1 != types.I32 {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only div 2 i32 values(\"%s\" is \"%s\" not i32)\n", words[idx-1].Str, t1.Readable())
        fmt.Fprintln(os.Stderr, "\t" + words[idx-1].At())
        os.Exit(1)
    }
    if t2 != types.I32 {
        fmt.Fprintf(os.Stderr, "[ERROR] you can only div 2 i32 values(\"%s\" is \"%s\" not i32)\n", words[idx+1].Str, t2.Readable())
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    op := Op{ Type: OP_DIV, Token: words[idx], Operants: []string{ words[idx-4].Str, words[idx+1].Str } }
    Ops = append(Ops, op)

    return idx + 1
}

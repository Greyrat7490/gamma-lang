package prs

import (
    "fmt"
    "os"
)

func prsDefFn(words []Token, idx int) int {
    if words[idx+1].Str == "main" {
        isMainDefined = true
    }

    var op Op = Op{ Type: OP_DEF_FN, Token: words[idx], Operants: []string{ words[idx+1].Str } }
    Ops = append(Ops, op)

    idx = prsDecArgs(words, idx)

    for ; idx < len(words); idx++ {
        switch words[idx].Str {
        case "var":
            idx = prsDecVar(words, idx)
        case ":=":
            idx = prsDefVar(words, idx)
        case "+":
            idx = prsAdd(tokens, idx)
        case "-":
            idx = prsSub(tokens, idx)
        case "*":
            idx = prsMul(tokens, idx)
        case "/":
            idx = prsDiv(tokens, idx)
        case "fn":
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
            os.Exit(1)
        case "}":
            prsEnd(words, idx)
            return idx
        default:
            idx = prsCallFn(words, idx)
        }
    }

    fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" was not closed (missing \"}\")\n", words[idx+1].Str)
    fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
    os.Exit(1)

    return -1
}

func prsEnd(words []Token, idx int) {
    Ops = append(Ops, Op{ Type: OP_END_FN, Token: words[idx] })
}

func prsDecArgs(words []Token, idx int) int {
    if len(words) < idx + 2 || words[idx+2].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].At())
        os.Exit(1)
    }

    var op Op = Op{ Type: OP_DEC_ARGS, Token: words[idx+3] }

    b := false
    for _, w := range words[idx+3:] {
        if w.Str == ")" {
            b = true
            break
        }

        if w.Str == "{" || w.Str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }
        op.Operants = append(op.Operants, w.Str)
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\" for function \"%s\"\n", words[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    if len(op.Operants) > 0 {
        Ops = append(Ops, op)
    }

    return idx + len(op.Operants) + 5
}

func prsCallFn(words []Token, idx int) int {
    var op Op = Op{ Type: OP_CALL_FN, Token: words[idx], Operants: []string{ words[idx].Str } }

    if len(words) < idx + 1 || words[idx+1].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    idx = prsDefArgs(words, idx)

    Ops = append(Ops, op)

    return idx
}

func prsDefArgs(words []Token, idx int) int {
    var op Op = Op{ Type: OP_DEF_ARGS, Token: words[idx+2], Operants: []string{ words[idx].Str } }

    b := false
    for _, w := range words[idx+2:] {
        if w.Str == ")" {
            b = true
            break
        }

        if w.Str == "{" || w.Str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }

        op.Operants = append(op.Operants, w.Str)
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    Ops = append(Ops, op)

    return idx + len(op.Operants) + 1
}

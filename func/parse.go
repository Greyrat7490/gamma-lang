package fn

import (
    "fmt"
    "gorec/types"
    "gorec/parser"
    "gorec/vars"
    "os"
)

type fnHead struct {
    name string
    args []arg
    col int
    line int
}

type arg struct {
    name string
    argType types.Type
}

func (f *fnHead) At() string {
    return fmt.Sprintf("at line: %d, col: %d", f.line, f.col)
}

var funcs []fnHead
var curFunc int = -1

var isMainDefined bool = false


func ParseDefine(words []prs.Token, idx int) int {
    if words[idx+1].Str == "main" {
        isMainDefined = true
    }

    var f fnHead = fnHead{
        name: words[idx+1].Str,
        col: words[idx+1].Col,
        line: words[idx+1].Line,
    }
    curFunc = len(funcs)
    funcs = append(funcs, f)

    var op prs.Op = prs.Op{ Type: prs.OP_DEF_FN, Token: words[idx], Operants: []string{ words[idx+1].Str } }
    prs.Ops = append(prs.Ops, op)

    idx = parseDeclareArgs(words, &funcs[curFunc], idx)

    if len(f.args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only accept one argument max at the moment")
        os.Exit(1)
    }

    for ; idx < len(words); idx++ {
        switch words[idx].Str {
        case "var":
            idx = vars.ParseDeclare(words, idx)
        case ":=":
            idx = vars.ParseDefine(words, idx)
        case "fn":
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
            os.Exit(1)
        case "}":
            parseEnd(words, idx)
            return idx
        default:
            idx = parseCallFunc(words, idx)
        }
    }

    return -1
}

func parseEnd(words []prs.Token, idx int) {
    var op prs.Op = prs.Op{ Type: prs.OP_END_FN, Token: words[idx] }

    f := get(funcs[curFunc].name)

    // TODO: later local variables
    for _, a := range f.args {
        op.Operants = append(op.Operants, a.name)
    }

    prs.Ops = append(prs.Ops, op)
    curFunc = -1
}

func parseDeclareArgs(words []prs.Token, f *fnHead, idx int) int {
    if len(words) < idx + 2 || words[idx+2].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].At())
        os.Exit(1)
    }

    var op prs.Op = prs.Op{ Type: prs.OP_DEC_ARGS, Token: words[idx] }

    name := ""
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

        if name == "" {
            name = w.Str
        } else {
            if t := types.ToType(w.Str); t != -1 {
                f.args = append(f.args, arg{ name: name, argType: t })
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] unknown type \"%s\"\n", w.Str)
                fmt.Fprintln(os.Stderr, "\t" + w.At())
                os.Exit(1)
            }
        }

        op.Operants = append(op.Operants, w.Str)
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\" for function \"%s\"\n", words[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    if len(op.Operants) > 0 {
        prs.Ops = append(prs.Ops, op)
    }

    return idx + len(op.Operants) + 5
}

func parseCallFunc(words []prs.Token, idx int) int {
    f := get(words[idx].Str)

    var op prs.Op = prs.Op{ Type: prs.OP_CALL_FN, Token: words[idx], Operants: []string{ words[idx].Str } }

    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] undeclared name \"%s\"\n", words[idx].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
        os.Exit(1)
    }

    if len(words) < idx + 1 || words[idx+1].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    idx = parseCallArgs(words, idx)

    prs.Ops = append(prs.Ops, op)

    return idx
}

func parseCallArgs(words []prs.Token, idx int) (nextIdx int) {
    var op prs.Op = prs.Op{ Type: prs.OP_DEF_ARGS, Token: words[idx], Operants: []string{ words[idx].Str } }

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

    prs.Ops = append(prs.Ops, op)

    return idx + len(op.Operants) + 1
}

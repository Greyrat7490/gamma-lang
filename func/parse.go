package function

import (
    "fmt"
    "strconv"
    "os"
    "gorec/types"
    "gorec/parser"
)

func parseCallFunc(words []prs.Word, idx int) (*Function, int) {
    f := getFunc(words[idx].Str)

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

    idx = parseCallArgs(words, f, idx)

    return f, idx
}

func parseCallArgs(words []prs.Word, f *Function, idx int) (nextIdx int) {
    argCount := 0
    b := false
    for i, w := range words[idx+2:] {
        if w.Str == ")" {
            b = true
            argCount = i
            break
        }

        if w.Str == "{" || w.Str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }

        t := f.Args[i].ArgType
        isVar := false
        if w.Str[0] == '"' && w.Str[len(w.Str) - 1] == '"' {
            t = types.Str
        } else if _, err := strconv.Atoi(w.Str); err == nil {
            t = types.I32
        } else {
            isVar = true
        }

        f.Args[i].IsVar = isVar
        f.Args[i].Value = w.Str

        if f.Args[i].ArgType != t {
            fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" expected as %d argument \"%s\" but got \"%s\"\n",
                f.Name, i, f.Args[i].ArgType.Readable(), t.Readable())
            fmt.Fprintln(os.Stderr, "\t" + w.At())
            os.Exit(1)
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    if argCount != len(f.Args) {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" expected %d arguments but got %d\n",
            f.Name, len(f.Args), argCount)
        fmt.Fprintln(os.Stderr, "\t" + f.At())
        os.Exit(1)
    }

    return idx + len(f.Args) + 2
}


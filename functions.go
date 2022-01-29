package main

import (
    "fmt"
    "os"
    "strconv"
)

type arg struct {
    isVar bool
    argType gType
    value int       // regIdx if isVar, strIdx if argType is str
}

var inMain bool = false

func defineEntry(asm *os.File, words []word, i int) int {
    if words[i+1].str != "main" {
        fmt.Fprintf(os.Stderr, "[ERROR] only \"main\" is allowed as name for the entry function (not \"%s\")\n", words[i+1].str)
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
        os.Exit(1)
    }

    asm.WriteString("main:\n")
    
    inMain = true
    
    return i + 4
}

func getArgs(words []word, expectedArgCount int) (args []arg) {
    if len(words) < 2 || words[1].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[1].at())
        os.Exit(1)
    }

    for _, w := range words[2:] {
        if w.str == ")" {
            break
        }

        if w.str[0] == '"' && w.str[len(w.str) - 1] == '"' {
            args = append(args, arg{false, str, len(strLits)})
            addStrLit(w)
        } else if i, err := strconv.Atoi(w.str); err == nil {
            args = append(args, arg{false, i32, i})
        } else {
            if v := getVar(w.str); v != nil {
                args = append(args, arg{true, v.vartype, v.regIdx})
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", w.str)
                fmt.Fprintln(os.Stderr, "\t" + w.at())
                os.Exit(1)
            }
        }
    }

    if len(words) - 2 == len(args) {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    if len(args) != expectedArgCount {
        fmt.Fprintf(os.Stderr, "[ERROR] function takes %d argument but got %d\n", expectedArgCount, len(args))
        fmt.Fprintln(os.Stderr, "\t" + words[0].at())
        os.Exit(1)
    }

    return args
}

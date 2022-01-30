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

type function struct {
    name string
}

var funcs []function
var curFunc int = -1

var mainDef bool = false

func defineFunc(asm *os.File, words []word, i int) int {
    if words[i+1].str == "main" {
        mainDef = true
    }

    asm.WriteString(words[i+1].str + ":\n")

    curFunc = len(funcs)
    funcs = append(funcs, function{words[i+1].str})

    for i += 5; i < len(words); i++ {
        switch words[i].str {
        case "var":
            i = declareVar(words, i)
        case ":=":
            i = defineVar(words, i)
        case "println":
            i = write(asm, words, i)
        case "exit":
            i = exit(asm, words, i)
        case "fn":
            fmt.Fprintln(os.Stderr, "[ERROR] you are ot allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
            os.Exit(1)
        case "}":
            endFunc(asm)
            return i
        default:
            i = callFunc(asm, words, i)
        }
    }

    return i
}

func endFunc(asm *os.File) {
    asm.WriteString("ret\n")
    curFunc = -1
}

func getFunc(funcName string) *function {
    for _, f := range funcs {
        if f.name == funcName {
            return &f
        }
    }

    return nil
}

func callFunc(asm *os.File, words []word, i int) int {
    if f := getFunc(words[i].str); f == nil {
        // TODO check for ()

        fmt.Fprintf(os.Stderr, "[ERROR] keyword \"%s\" is not supported\n", words[i].str)
        fmt.Fprintln(os.Stderr, "\t" + words[i].at())
        os.Exit(1)
    } else {
        asm.WriteString("call " + f.name + "\n")
    }

    return i + 2
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

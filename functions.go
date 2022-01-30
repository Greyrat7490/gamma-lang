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
    col int
    line int
}

func (f *function) at() string {
    return fmt.Sprintf("at line: %d, col: %d", f.line, f.col)
}

var funcs []function
var curFunc int = -1

var mainDef bool = false

func defineFunc(asm *os.File, words []word, i int) int {
    if words[i+1].str == "main" {
        mainDef = true
    }

    declareArgs(words, i) // checks for () later actually declare args

    asm.WriteString(words[i+1].str + ":\n")

    curFunc = len(funcs)
    funcs = append(funcs, function{
        name: words[i+1].str,
        col: words[i+1].col,
        line: words[i+1].line,
    })

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
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
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
        fmt.Fprintf(os.Stderr, "[ERROR] keyword \"%s\" is not supported\n", words[i].str)
        fmt.Fprintln(os.Stderr, "\t" + words[i].at())
        os.Exit(1)
        return 1
    } else {
        if len(words) < i + 1 || words[i+1].str != "(" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
            fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
            os.Exit(1)
        }
        args := defineArgs(words[i+1:])
        checkArgs(args, words[i+1], 0) // no args supported yet

        asm.WriteString("call " + f.name + "\n")
        return i + len(args) + 2
    }
}

func declareArgs(words []word, i int) {
    if len(words) < i + 2 || words[i+2].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[i+2].at())
        os.Exit(1)
    }

    b := false
    for _, w := range words[i+3:] {
        if w.str == ")" {
            b = true
            break
        }

        if w.str == "{" || w.str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.at())
            os.Exit(1)
        }

        // TODO
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }
}

func defineArgs(words []word) (args []arg) {
    if len(words) < 1 || words[0].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[0].at())
        os.Exit(1)
    }

    b := false
    for _, w := range words[1:] {
        if w.str == ")" {
            b = true
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

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    return args
}

func checkArgs(args []arg, w word, expected int) {
    if len(args) != expected {
        fmt.Fprintf(os.Stderr, "[ERROR] function takes %d argument but got %d\n", expected, len(args))
        fmt.Fprintln(os.Stderr, "\t" + w.at())
        os.Exit(1)
    }

    // TODO: check types
}

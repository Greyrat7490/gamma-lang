package main

import (
    "fmt"
    "os"
    "strconv"
)

// calling convention (temporary):
// - one argument max
// - i32 -> r9 = num
// - str -> r9 = addr, r10 = size
// TODO: C calling convention

type arg struct {
    name string
    isVar bool
    argType gType
    value string
}

type function struct {
    name string
    args []arg
    col int
    line int
}

func (f *function) at() string {
    return fmt.Sprintf("at line: %d, col: %d", f.line, f.col)
}

var funcs []function
var curFunc int = -1

var mainDef bool = false

func defineFunc(asm *os.File, words []word, idx int) int {
    if words[idx+1].str == "main" {
        mainDef = true
    }

    args, nextIdx := declareArgs(words, idx)

    if len(args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only acept one argument max at the moment")
        os.Exit(1)
    }

    curFunc = len(funcs)
    funcs = append(funcs, function{
        name: words[idx+1].str,
        args: args,
        col: words[idx+1].col,
        line: words[idx+1].line,
    })

    asm.WriteString(words[idx+1].str + ":\n")
    for idx = nextIdx; idx < len(words); idx++ {
        switch words[idx].str {
        case "var":
            idx = declareVar(words, idx)
        case ":=":
            idx = defineVar(words, idx)
        case "fn":
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + words[idx].at())
            os.Exit(1)
        case "}":
            endFunc(asm)
            return idx
        default:
            f, nextIdx := parseCallFunc(words, idx)
            idx = nextIdx
            callFunc(asm, f)
        }
    }

    return idx
}

func endFunc(asm *os.File) {
    asm.WriteString("ret\n\n")

    for _, a := range funcs[curFunc].args {
        rmVar(a.name)
    }

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

func parseCallFunc(words []word, idx int) (*function, int) {
    f := getFunc(words[idx].str)

    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] undeclared name \"%s\"\n", words[idx].str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx].at())
        os.Exit(1)
    }

    if len(words) < idx + 1 || words[idx+1].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].at())
        os.Exit(1)
    }

    idx = parseCallArgs(words, f, idx)

    return f, idx
}

func parseCallArgs(words []word, f *function, idx int) (nextIdx int) {
    argCount := 0
    b := false
    for i, w := range words[idx+2:] {
        if w.str == ")" {
            b = true
            argCount = i
            break
        }

        if w.str == "{" || w.str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.at())
            os.Exit(1)
        }

        t := f.args[i].argType
        isVar := false
        if w.str[0] == '"' && w.str[len(w.str) - 1] == '"' {
            t = str
        } else if _, err := strconv.Atoi(w.str); err == nil {
            t = i32
        } else {
            isVar = true
        }

        f.args[i].isVar = isVar
        f.args[i].value = w.str

        if f.args[i].argType != t {
            fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" expected as %d argument \"%s\" but got \"%s\"\n",
                f.name, i, f.args[i].argType.readable(), t.readable())
            fmt.Fprintln(os.Stderr, "\t" + w.at())
            os.Exit(1)
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    if argCount != len(f.args) {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" expected %d arguments but got %d\n",
            f.name, len(f.args), argCount)
        fmt.Fprintln(os.Stderr, "\t" + f.at())
        os.Exit(1)
    }

    return idx + len(f.args) + 2
}

func callFunc(asm *os.File, f *function) {
    defineArgs(asm, f)

    asm.WriteString("call " + f.name + "\n")
}

func declareArgs(words []word, idx int) (args []arg, nextIdx int) {
    if len(words) < idx + 2 || words[idx+2].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].at())
        os.Exit(1)
    }

    var a arg
    argName := true

    b := false
    for _, w := range words[idx+3:] {
        if w.str == ")" {
            b = true
            break
        }

        if w.str == "{" || w.str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.at())
            os.Exit(1)
        }

        if argName {
            a.name = w.str
            argName = false
        } else {
            a.argType = toType(w.str)
            argName = true

            args = append(args, a)

            // see calling convention
            // 5 = r9, 6 = r10
            var regs []int
            switch a.argType {
            case str:
                regs = []int { 5, 6 }
            case i32:
                regs = []int { 5 }
            default:
                fmt.Fprintf(os.Stderr, "[ERROR] unknown type \"%s\"\n", w.str)
                fmt.Fprintln(os.Stderr, "\t" + w.at())
                os.Exit(1)
            }

            vars = append(vars, variable{a.name, regs, a.argType})
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\" for function \"%s\"\n", words[idx+1].str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].at())
        os.Exit(1)
    }

    return args, idx + len(args) * 2 + 5
}

func defineArgs(asm *os.File, f *function) {
    for i, a := range f.args {
        if otherVar := getVar(a.value); otherVar != nil {
            if otherVar.vartype != a.argType {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                    f.name, i, a.argType.readable(), otherVar.vartype.readable())
                fmt.Fprintln(os.Stderr, "\t" + f.at())
                os.Exit(1)
            }

            // skip if r9 is already set correct
            if otherVar.regs[0] == 5 {
                return
            }

            switch a.argType {
            case str:
                asm.WriteString(fmt.Sprintf("mov r9, %s\n", registers[otherVar.regs[0]].name))
                asm.WriteString(fmt.Sprintf("mov r10, %s\n", registers[otherVar.regs[1]].name))

            case i32:
                asm.WriteString(fmt.Sprintf("mov r9, %s\n", registers[otherVar.regs[0]].name))

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) function.go defineArgs()")
                os.Exit(1)
            }
        } else {
            switch a.argType {
            case str:
                asm.WriteString(fmt.Sprintf("mov r9, str%d\n", len(strLits)))
                addStrLit(a.value)
                asm.WriteString(fmt.Sprintf("mov r10, %d\n", strLits[len(strLits)-1].size))

            case i32:
                i, _ := strconv.Atoi(a.value)
                asm.WriteString(fmt.Sprintf("mov r9, %d\n", i))

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) function.go defineArgs()")
                os.Exit(1)
            }
        }
    }
}

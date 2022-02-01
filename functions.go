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

func defineFunc(asm *os.File, words []word, i int) int {
    if words[i+1].str == "main" {
        mainDef = true
    }

    args, nextIdx := declareArgs(words, i)

    if len(args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only acept one argument max at the moment")
        os.Exit(1)
    }

    curFunc = len(funcs)
    funcs = append(funcs, function{
        name: words[i+1].str,
        args: args,
        col: words[i+1].col,
        line: words[i+1].line,
    })

    asm.WriteString(words[i+1].str + ":\n")
    for i = nextIdx; i < len(words); i++ {
        switch words[i].str {
        case "var":
            i = declareVar(words, i)
        case ":=":
            i = defineVar(words, i)
        case "fn":
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
            os.Exit(1)
        case "}":
            endFunc(asm)
            return i
        default:
            f, nextIdx := parseCallFunc(words, i)
            i = nextIdx
            callFunc(asm, f)
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

func parseCallFunc(words []word, i int) (*function, int) {
    f := getFunc(words[i].str)

    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] keyword \"%s\" is not supported\n", words[i].str)
        fmt.Fprintln(os.Stderr, "\t" + words[i].at())
        os.Exit(1)
    }

    if len(words) < i + 1 || words[i+1].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
        os.Exit(1)
    }

    i = parseCallArgs(words, f, i)

    return f, i
}

func parseCallArgs(words []word, f *function, i int) (nextIdx int) {
    argCount := 0
    b := false
    for ai, w := range words[i+2:] {
        if w.str == ")" {
            b = true
            argCount = ai
            break
        }

        if w.str == "{" || w.str == "}" {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \")\"")
            fmt.Fprintln(os.Stderr, "\t" + w.at())
            os.Exit(1)
        }

        t := f.args[ai].argType
        isVar := false
        if w.str[0] == '"' && w.str[len(w.str) - 1] == '"' {
            t = str
        } else if _, err := strconv.Atoi(w.str); err == nil {
            t = i32
        } else {
            isVar = true
        }

        f.args[ai].isVar = isVar
        f.args[ai].value = w.str

        // fmt.Printf("%s = %s\n", f.args[ai].name, w.str)

        if f.args[ai].argType != t {
            fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" expected as %d argument \"%s\" but got \"%s\"\n",
                f.name, ai, f.args[ai].argType.readable(), t.readable())
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

    return i + len(f.args) + 2
}

func callFunc(asm *os.File, f *function) {
    defineArgs(asm, f)

    asm.WriteString("call " + f.name + "\n")

    clearArgs(asm, f)
}

func clearArgs(asm *os.File, f *function) {
    if f.args[0].argType == i32 {
        asm.WriteString("pop r9\n")
    } else {
        asm.WriteString("pop r10\n")
        asm.WriteString("pop r9\n")
    }

    rmVar(f.args[0].name)
}

func declareArgs(words []word, i int) (args []arg, nextIdx int) {
    if len(words) < i + 2 || words[i+2].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[i+2].at())
        os.Exit(1)
    }

    var a arg
    argName := true

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

        if argName {
            a.name = w.str
            argName = false
        } else {
            a.argType = toType(w.str)
            argName = true

            args = append(args, a)
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    return args, i + len(args) * 2 + 5
}

func defineArgs(asm *os.File, f *function) {
    if f.args[0].argType == i32 {
        asm.WriteString("push r9\n")
    } else {
        asm.WriteString("push r9\n")
        asm.WriteString("push r10\n")
    }

    for _, a := range f.args {
        v := variable{a.name, len(vars), a.argType, -1}
        vars = append(vars, v)
        if a.isVar {
            // works for now (except global variables)
            // does not check the variable name or type
            // TODO: get register of the variable
        } else {
            switch a.argType {
            case str:
                registers[v.regIdx].isAddr = true;
                registers[v.regIdx].value = len(strLits);

                asm.WriteString(fmt.Sprintf("mov r9, str%d\n", registers[v.regIdx].value))

                addStrLit(a.value)

                asm.WriteString(fmt.Sprintf("mov r10, %d\n", strLits[len(strLits)-1].size))

            case i32:
                i, _ := strconv.Atoi(a.value)

                registers[v.regIdx].isAddr = false;
                registers[v.regIdx].value = i;
                asm.WriteString(fmt.Sprintf("mov r9, %d\n", i))

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) function.go defineArgs()")
                os.Exit(1)
            }
        }
    }
}

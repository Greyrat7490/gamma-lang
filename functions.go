package main

import (
    "fmt"
    "os"
    "strconv"
)

type argDec struct {
    name string
    argType gType
}

type function struct {
    name string
    args []argDec
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

        defineArgs(asm, words[i+1:], f)

        asm.WriteString("call " + f.name + "\n")
        return i + len(f.args) + 2
    }
}

func declareArgs(words []word, i int) (args []argDec, nextIdx int) {
    if len(words) < i + 2 || words[i+2].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[i+2].at())
        os.Exit(1)
    }

    var a argDec
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
            v := variable{a.name, len(vars), a.argType, -1}
            vars = append(vars, v)
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    return args, i + len(args) * 2 + 5
}

func defineArgs(asm *os.File, words []word, f *function) {
    if len(words) < 1 || words[0].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[0].at())
        os.Exit(1)
    }

    b := false
    for i, w := range words[1:] {
        if w.str == ")" {
            b = true
            break
        }

        a := f.args[i]
        v := getVar(a.name)

        if isLit(w.str) {
            switch a.argType {
            case str:
                // later push registers
                registers[v.regIdx].isAddr = true;
                registers[v.regIdx].value = len(strLits);
                asm.WriteString(fmt.Sprintf("mov %s, str%d\n", registers[v.regIdx].name, registers[v.regIdx].value))

                addStrLit(w)

            case i32:
                i, _ := strconv.Atoi(w.str)

                // later push registers
                registers[v.regIdx].isAddr = false;
                registers[v.regIdx].value = i;
                asm.WriteString(fmt.Sprintf("mov %s, %d\n", registers[v.regIdx].name, i))

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) function.go defineArgs()")
                os.Exit(1)
            }
        } else {
            // TODO: check if var is defined
            if otherVar := getVar(w.str); otherVar != nil {
                // later push registers
                registers[v.regIdx].isAddr = registers[otherVar.regIdx].isAddr;
                registers[v.regIdx].value = registers[otherVar.regIdx].value;
                asm.WriteString(fmt.Sprintf("mov %s, %s\n", registers[v.regIdx].name, registers[otherVar.regIdx].name))
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
}

func checkArgs(w word, f *function) {
    // TODO: check types
}

package fn

import (
    "fmt"
    "os"
    "strconv"
    "gorec/types"
    "gorec/parser"
    "gorec/vars"
    "gorec/str"
)

// calling convention (temporary):
// - one argument max
// - i32 -> r9 = num
// - str -> r9 = addr, r10 = size
// TODO: C calling convention

type Func struct {
    Name string
    Args []arg
    Col int
    Line int
}

type arg struct {
    name string
    isVar bool
    argType types.Type
    value string
}

var funcs []Func
var curFunc int = -1

var mainDef bool = false


func (f *Func) At() string {
    return fmt.Sprintf("at line: %d, col: %d", f.Line, f.Col)
}

func Define(asm *os.File, words []prs.Word, idx int) int {
    if words[idx+1].Str == "main" {
        mainDef = true
    }

    args, nextIdx := declareArgs(words, idx)

    if len(args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only acept one argument max at the moment")
        os.Exit(1)
    }

    curFunc = len(funcs)
    funcs = append(funcs, Func{
        Name: words[idx+1].Str,
        Args: args,
        Col: words[idx+1].Col,
        Line: words[idx+1].Line,
    })

    asm.WriteString(words[idx+1].Str + ":\n")
    for idx = nextIdx; idx < len(words); idx++ {
        switch words[idx].Str {
        case "var":
            idx = vars.Declare(words, idx)
        case ":=":
            idx = vars.Define(words, idx)
        case "fn":
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
            os.Exit(1)
        case "}":
            end(asm)
            return idx
        default:
            f, nextIdx := parseCallFunc(words, idx)
            idx = nextIdx
            callFunc(asm, f)
        }
    }

    return idx
}

func end(asm *os.File) {
    asm.WriteString("ret\n\n")

    for _, a := range funcs[curFunc].Args {
        vars.Remove(a.name)
    }

    curFunc = -1
}

func get(funcName string) *Func {
    for _, f := range funcs {
        if f.Name == funcName {
            return &f
        }
    }

    return nil
}
func callFunc(asm *os.File, f *Func) {
    defineArgs(asm, f)

    asm.WriteString("call " + f.Name + "\n")
}

func declareArgs(words []prs.Word, idx int) (args []arg, nextIdx int) {
    if len(words) < idx + 2 || words[idx+2].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].At())
        os.Exit(1)
    }

    var a arg
    argName := true

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

        if argName {
            a.name = w.Str
            argName = false
        } else {
            a.argType = types.ToType(w.Str)
            argName = true

            args = append(args, a)

            // see calling convention
            // 5 = r9, 6 = r10
            var regs []int
            switch a.argType {
            case types.Str:
                regs = []int { 5, 6 }
            case types.I32:
                regs = []int { 5 }
            default:
                fmt.Fprintf(os.Stderr, "[ERROR] unknown type \"%s\"\n", w.Str)
                fmt.Fprintln(os.Stderr, "\t" + w.At())
                os.Exit(1)
            }

            vars.Add(vars.Var{Name: a.name, Regs: regs, Vartype: a.argType})
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\" for function \"%s\"\n", words[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    return args, idx + len(args) * 2 + 5
}

func defineArgs(asm *os.File, f *Func) {
    for i, a := range f.Args {
        if otherVar := vars.Get(a.value); otherVar != nil {
            if otherVar.Vartype != a.argType {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                    f.Name, i, a.argType.Readable(), otherVar.Vartype.Readable())
                fmt.Fprintln(os.Stderr, "\t" + f.At())
                os.Exit(1)
            }

            // skip if r9 is already set correct
            if otherVar.Regs[0] == 5 {
                return
            }

            switch a.argType {
            case types.Str:
                asm.WriteString(fmt.Sprintf("mov r9, %s\n", vars.Registers[otherVar.Regs[0]].Name))
                asm.WriteString(fmt.Sprintf("mov r10, %s\n", vars.Registers[otherVar.Regs[1]].Name))

            case types.I32:
                asm.WriteString(fmt.Sprintf("mov r9, %s\n", vars.Registers[otherVar.Regs[0]].Name))

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) function.go defineArgs()")
                os.Exit(1)
            }
        } else {
            switch a.argType {
            case types.Str:
                strIdx := str.Add(a.value)
                asm.WriteString(fmt.Sprintf("mov r9, str%d\n", strIdx))
                asm.WriteString(fmt.Sprintf("mov r10, %d\n", str.GetSize(strIdx)))

            case types.I32:
                i, _ := strconv.Atoi(a.value)
                asm.WriteString(fmt.Sprintf("mov r9, %d\n", i))

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) function.go defineArgs()")
                os.Exit(1)
            }
        }
    }
}

func Add(f *Func) {
    funcs = append(funcs, *f)
}

func AddBuildIn(name string, argname string, argtype types.Type) {
    funcs = append(funcs, Func{Name: name, Args: []arg{{name: argname, argType: argtype}}})
}

func Checks() {
    if !mainDef {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }

    if curFunc != -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" was not closed (missing \"}\")\n", funcs[curFunc].Name)
        fmt.Fprintln(os.Stderr, "\t" + funcs[curFunc].At())
        os.Exit(1)
    }
}

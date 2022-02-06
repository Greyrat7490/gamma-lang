package function

import (
    "fmt"
    "os"
    "strconv"
    "gorec/types"
    "gorec/parser"
    "gorec/vars"
)

// calling convention (temporary):
// - one argument max
// - i32 -> r9 = num
// - str -> r9 = addr, r10 = size
// TODO: C calling convention

type Arg struct {
    Name string
    IsVar bool
    ArgType types.Type
    Value string
}

type Function struct {
    Name string
    Args []Arg
    Col int
    Line int
}

func (f *Function) At() string {
    return fmt.Sprintf("at line: %d, col: %d", f.Line, f.Col)
}

var funcs []Function
var curFunc int = -1

var mainDef bool = false

func DefineFunc(asm *os.File, words []prs.Word, idx int) int {
    if words[idx+1].Str == "main" {
        mainDef = true
    }

    args, nextIdx := declareArgs(words, idx)

    if len(args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only acept one argument max at the moment")
        os.Exit(1)
    }

    curFunc = len(funcs)
    funcs = append(funcs, Function{
        Name: words[idx+1].Str,
        Args: args,
        Col: words[idx+1].Col,
        Line: words[idx+1].Line,
    })

    asm.WriteString(words[idx+1].Str + ":\n")
    for idx = nextIdx; idx < len(words); idx++ {
        switch words[idx].Str {
        case "var":
            idx = vars.DeclareVar(words, idx)
        case ":=":
            idx = vars.DefineVar(words, idx)
        case "fn":
            fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
            fmt.Fprintln(os.Stderr, "\t" + words[idx].At())
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

    for _, a := range funcs[curFunc].Args {
        vars.RmVar(a.Name)
    }

    curFunc = -1
}

func getFunc(funcName string) *Function {
    for _, f := range funcs {
        if f.Name == funcName {
            return &f
        }
    }

    return nil
}
func callFunc(asm *os.File, f *Function) {
    defineArgs(asm, f)

    asm.WriteString("call " + f.Name + "\n")
}

func declareArgs(words []prs.Word, idx int) (args []Arg, nextIdx int) {
    if len(words) < idx + 2 || words[idx+2].Str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[idx+2].At())
        os.Exit(1)
    }

    var a Arg
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
            a.Name = w.Str
            argName = false
        } else {
            a.ArgType = types.ToType(w.Str)
            argName = true

            args = append(args, a)

            // see calling convention
            // 5 = r9, 6 = r10
            var regs []int
            switch a.ArgType {
            case types.Str:
                regs = []int { 5, 6 }
            case types.I32:
                regs = []int { 5 }
            default:
                fmt.Fprintf(os.Stderr, "[ERROR] unknown type \"%s\"\n", w.Str)
                fmt.Fprintln(os.Stderr, "\t" + w.At())
                os.Exit(1)
            }

            vars.AddVar(vars.Variable{a.Name, regs, a.ArgType})
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\" for function \"%s\"\n", words[idx+1].Str)
        fmt.Fprintln(os.Stderr, "\t" + words[idx+1].At())
        os.Exit(1)
    }

    return args, idx + len(args) * 2 + 5
}

func defineArgs(asm *os.File, f *Function) {
    for i, a := range f.Args {
        if otherVar := vars.GetVar(a.Value); otherVar != nil {
            if otherVar.Vartype != a.ArgType {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                    f.Name, i, a.ArgType.Readable(), otherVar.Vartype.Readable())
                fmt.Fprintln(os.Stderr, "\t" + f.At())
                os.Exit(1)
            }

            // skip if r9 is already set correct
            if otherVar.Regs[0] == 5 {
                return
            }

            switch a.ArgType {
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
            switch a.ArgType {
            case types.Str:
                asm.WriteString(fmt.Sprintf("mov r9, str%d\n", len(types.StrLits)))
                types.AddStrLit(a.Value)
                asm.WriteString(fmt.Sprintf("mov r10, %d\n", types.StrLits[len(types.StrLits)-1].Size))

            case types.I32:
                i, _ := strconv.Atoi(a.Value)
                asm.WriteString(fmt.Sprintf("mov r9, %d\n", i))

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) function.go defineArgs()")
                os.Exit(1)
            }
        }
    }
}

func AddFunc(f *Function) {
    funcs = append(funcs, *f)
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

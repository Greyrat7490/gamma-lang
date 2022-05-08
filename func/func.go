package fn

import (
    "os"
    "fmt"
    "strconv"
    "gorec/str"
    "gorec/vars"
    "gorec/token"
    "gorec/types"
    "gorec/conditions"
)

// calling convention (temporary):
// - one argument max
// - i32 -> r10 = num
// - str -> r10 = addr, r11 = size
// TODO: C calling convention

var funcs []fnHead
var curFunc int = -1


type fnHead struct {
    name token.Token
    args []Arg
}

type Arg struct {
    Name string
    Type types.Type
}

func (a Arg) String() string {
    return fmt.Sprintf("{%s(Name) %s(Typename)}", a.Name, a.Type.Readable())
}

func (f *fnHead) At() string {
    return f.name.At()
}

func GetFn(funcName string) *fnHead {
    for _, f := range funcs {
        if f.name.Str == funcName {
            return &f
        }
    }

    return nil
}


func Define(asm *os.File, fnName token.Token) {
    var f fnHead = fnHead{
        name: fnName,
    }
    curFunc = len(funcs)
    funcs = append(funcs, f)
    vars.IsGlobalScope = false

    asm.WriteString(fnName.Str + ":\n")
}

func End(asm *os.File) {
    f := funcs[curFunc]

    // TODO: later local variables
    for _, a := range f.args {
        vars.Remove(a.Name)
    }
    curFunc = -1
    vars.IsGlobalScope = true

    cond.ResetCount()

    asm.WriteString("ret\n\n")
}

func CallFunc(asm *os.File, fnName token.Token) {
    if f := GetFn(fnName.Str); f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] undeclared name \"%s\"\n", fnName.Str)
        fmt.Fprintln(os.Stderr, "\t" + fnName.At())
        os.Exit(1)
    }

    asm.WriteString("call " + fnName.Str + "\n")
}

func DeclareArgs(args []Arg) {
    if curFunc == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) CurFunc should be set")
        os.Exit(1)
    }

    f := &funcs[curFunc]

    for _, a := range args {
        f.args = append(f.args, Arg{ Name: a.Name, Type: a.Type })

        // see calling convention
        // 6 = r10, 7 = r11
        var regs []int
        const _ uint = 3 - types.TypesCount
        switch a.Type {
        case types.Str:
            regs = []int { 6, 7 }
        case types.I32, types.Bool:
            regs = []int { 6 }
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) TODO")
            os.Exit(1)
        }

        vars.Add(vars.Var{Name: a.Name, Regs: regs, Vartype: a.Type})
    }

    if len(f.args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only accept one argument max at the moment")
        fmt.Fprintln(os.Stderr, "\t" + f.At())
        os.Exit(1)
    }
}

func DefineArgByValue(asm *os.File, fnName token.Token, argNum int, value token.Token) {
    if f := GetFn(fnName.Str); f != nil {
       if t := types.TypeOfVal(value.Str); t != -1 {
            if t != f.args[argNum].Type {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                    f.name.Str, argNum, f.args[argNum].Type.Readable(), t.Readable())
                os.Exit(1)
            }

            const _ uint = 3 - types.TypesCount
            switch t {
            case types.Str:
                strIdx := str.Add(value.Str)
                asm.WriteString(fmt.Sprintf("mov r10, str%d\n", strIdx))
                asm.WriteString(fmt.Sprintf("mov r11, %d\n", str.GetSize(strIdx)))

            case types.I32:
                i, _ := strconv.Atoi(value.Str)
                asm.WriteString(fmt.Sprintf("mov r10, %d\n", i))

            case types.Bool:
                if value.Str == "true" {
                    asm.WriteString(fmt.Sprintf("mov r10, %d\n", 1))
                } else {
                    asm.WriteString(fmt.Sprintf("mov r10, %d\n", 0))
                }

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] could not get type of value \"%s\"\n", value.Str)
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", value.Str)
            fmt.Fprintln(os.Stderr, "\t" + fnName.At())
            os.Exit(1)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName.Str)
        os.Exit(1)
    }
}

func DefineArgByVar(asm *os.File, fnName token.Token, argNum int, varname token.Token) {
    if f := GetFn(fnName.Str); f != nil {
        if otherVar := vars.GetVar(varname.Str); otherVar != nil {
            if otherVar.Vartype != f.args[argNum].Type {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                    f.name.Str, argNum, f.args[argNum].Type.Readable(), otherVar.Vartype.Readable())
                fmt.Fprintln(os.Stderr, "\t" + fnName.At())
                os.Exit(1)
            }

            // skip if r10 is already set correct
            if otherVar.Regs[0] == 5 {
                return
            }

            const _ uint = 3 - types.TypesCount
            switch otherVar.Vartype {
            case types.Str:
                asm.WriteString(fmt.Sprintf("mov r10, %s\n", vars.Registers[otherVar.Regs[0]].Name))
                asm.WriteString(fmt.Sprintf("mov r11, %s\n", vars.Registers[otherVar.Regs[1]].Name))

            case types.I32, types.Bool:
                asm.WriteString(fmt.Sprintf("mov r10, %s\n", vars.Registers[otherVar.Regs[0]].Name))

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) type of var \"%s\" is not correct\n", otherVar.Name)
                os.Exit(1)
            }
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName.Str)
        os.Exit(1)
    }
}

func DefineArgByReg(asm *os.File, fnName token.Token, argNum int, reg string) {
    if f := GetFn(fnName.Str); f != nil {
        if f.args[argNum].Type == types.I32 {
            asm.WriteString(fmt.Sprintf("mov r10, %s\n", reg))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] expected function \"%s\" arg%d to be an i32 but got %v\n", fnName.Str, argNum, f.args[argNum].Type)
            os.Exit(1)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName.Str)
        os.Exit(1)
    }
}

func AddBuildIn(name string, argname string, argtype types.Type) {
    funcs = append(funcs, fnHead{
        name: token.Token{
            Str: name,
            Pos: token.Pos{Col: -1, Line: -1},
        },
        args: []Arg{{
            Name: argname, Type: argtype,
        }},
    })
}

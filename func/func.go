package fn

import (
    "fmt"
    "gorec/str"
    "gorec/types"
    "gorec/vars"
    "os"
    "strconv"
)

// calling convention (temporary):
// - one argument max
// - i32 -> r9 = num
// - str -> r9 = addr, r10 = size
// TODO: C calling convention

var funcs []fnHead
var curFunc int = -1


type fnHead struct {
    name string
    args []Arg
}

type Arg struct {
    Name string
    Type types.Type
}

func GetFn(funcName string) *fnHead {
    for _, f := range funcs {
        if f.name == funcName {
            return &f
        }
    }

    return nil
}


func Define(asm *os.File, fnName string) {
    var f fnHead = fnHead{
        name: fnName,
    }
    curFunc = len(funcs)
    funcs = append(funcs, f)

    asm.WriteString(fnName + ":\n")
}

func End(asm *os.File) {
    f := funcs[curFunc]

    // TODO: later local variables
    for _, a := range f.args {
        vars.Remove(a.Name)
    }
    curFunc = -1

    asm.WriteString("ret\n\n")
}

func CallFunc(asm *os.File, fnName string) {
    if f := GetFn(fnName); f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] undeclared name \"%s\"\n", fnName)
        os.Exit(1)
    }

    asm.WriteString("call " + fnName + "\n")
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
        // 5 = r9, 6 = r10
        var regs []int
        const _ uint = 2 - types.TypesCount
        switch a.Type {
        case types.Str:
            regs = []int { 5, 6 }
        case types.I32:
            regs = []int { 5 }
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) TODO")
            os.Exit(1)
        }

        vars.Add(vars.Var{Name: a.Name, Regs: regs, Vartype: a.Type})
    }

    if len(f.args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only accept one argument max at the moment")
        os.Exit(1)
    }
}

func DefineArgs(asm *os.File, fnName string, values []string) {
    if f := GetFn(fnName); f != nil {
        for i, val := range values {
            if otherVar := vars.GetVar(val); otherVar != nil {
                if otherVar.Vartype != f.args[i].Type {
                    fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                        f.name, i, f.args[i].Type.Readable(), otherVar.Vartype.Readable())
                    os.Exit(1)
                }

                // skip if r9 is already set correct
                if otherVar.Regs[0] == 5 {
                    return
                }

                const _ uint = 2 - types.TypesCount
                switch otherVar.Vartype {
                case types.Str:
                    asm.WriteString(fmt.Sprintf("mov r9, %s\n", vars.Registers[otherVar.Regs[0]].Name))
                    asm.WriteString(fmt.Sprintf("mov r10, %s\n", vars.Registers[otherVar.Regs[1]].Name))

                case types.I32:
                    asm.WriteString(fmt.Sprintf("mov r9, %s\n", vars.Registers[otherVar.Regs[0]].Name))

                default:
                    fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) type of var \"%s\" is not correct\n", otherVar.Name)
                    os.Exit(1)
                }
            } else {
                t := types.TypeOfVal(val)

                if t != f.args[i].Type {
                    fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                        f.name, i, f.args[i].Type.Readable(), t.Readable())
                    os.Exit(1)
                }

                const _ uint = 2 - types.TypesCount
                switch t {
                case types.Str:
                    strIdx := str.Add(val)
                    asm.WriteString(fmt.Sprintf("mov r9, str%d\n", strIdx))
                    asm.WriteString(fmt.Sprintf("mov r10, %d\n", str.GetSize(strIdx)))

                case types.I32:
                    i, _ := strconv.Atoi(val)
                    asm.WriteString(fmt.Sprintf("mov r9, %d\n", i))

                default:
                    fmt.Fprintf(os.Stderr, "[ERROR] could not get type of value \"%s\"\n", val)
                    os.Exit(1)
                }
            }
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName)
        os.Exit(1)
    }
}

func AddBuildIn(name string, argname string, argtype types.Type) {
    funcs = append(funcs, fnHead{name: name, args: []Arg{{Name: argname, Type: argtype}}})
}

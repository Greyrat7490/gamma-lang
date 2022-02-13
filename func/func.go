package fn

import (
    "fmt"
    "gorec/parser"
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

func Get(funcName string) *fnHead {
    for _, f := range funcs {
        if f.name == funcName {
            return &f
        }
    }

    return nil
}

func Define(asm *os.File, op *prs.Op) {
    f := get(op.Operants[0])
    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }

    asm.WriteString(op.Operants[0] + ":\n")
}

func End(asm *os.File, op *prs.Op) {
    for _, a := range op.Operants {
        vars.Remove(a)
    }

    curFunc = -1

    asm.WriteString("ret\n\n")
}

func get(funcName string) *fnHead {
    for _, f := range funcs {
        if f.name == funcName {
            return &f
        }
    }

    return nil
}
func CallFunc(asm *os.File, op *prs.Op) {
    f := get(op.Operants[0])

    asm.WriteString("call " + f.name + "\n")
}

func DeclareArgs(op *prs.Op) {
    isType := false
    var vartype types.Type
    var name string

    for _, a := range op.Operants {
        if isType {
            isType = false

            // see calling convention
            // 5 = r9, 6 = r10
            var regs []int
            switch a {
            case "str":
                regs = []int { 5, 6 }
                vartype = types.Str

            case "i32":
                regs = []int { 5 }
                vartype = types.I32

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) unknown type \"%s\"", a)
                os.Exit(1)
            }

            vars.Add(vars.Var{Name: name, Regs: regs, Vartype: vartype})
        } else {
            isType = true
            name = a
        }
    }
}

func DefineArgs(asm *os.File, op *prs.Op) {
    f := get(op.Operants[0])

    for i, val := range op.Operants[1:] {
        // TODO: check type of arg
        if otherVar := vars.Get(val); otherVar != nil {
            if otherVar.Vartype != f.args[i].argType {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                    f.name, i, f.args[i].argType.Readable(), otherVar.Vartype.Readable())
                fmt.Fprintln(os.Stderr, "\t" + f.At())
                os.Exit(1)
            }

            // skip if r9 is already set correct
            if otherVar.Regs[0] == 5 {
                return
            }

            switch otherVar.Vartype {
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
            switch types.TypeOfVal(val) {
            case types.Str:
                strIdx := str.Add(val)
                asm.WriteString(fmt.Sprintf("mov r9, str%d\n", strIdx))
                asm.WriteString(fmt.Sprintf("mov r10, %d\n", str.GetSize(strIdx)))

            case types.I32:
                i, _ := strconv.Atoi(val)
                asm.WriteString(fmt.Sprintf("mov r9, %d\n", i))

            default:
                fmt.Fprintln(os.Stderr, "[ERROR] (unreachable) function.go defineArgs()")
                os.Exit(1)
            }
        }
    }
}

func Add(f *fnHead) {
    funcs = append(funcs, *f)
}

func AddBuildIn(name string, argname string, argtype types.Type) {
    funcs = append(funcs, fnHead{name: name, args: []arg{{name: argname, argType: argtype}}})
}

func Checks() {
    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }

    if curFunc != -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" was not closed (missing \"}\")\n", funcs[curFunc].name)
        fmt.Fprintln(os.Stderr, "\t" + funcs[curFunc].At())
        os.Exit(1)
    }
}

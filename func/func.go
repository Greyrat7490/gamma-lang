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

var funcs []fnHead
var curFunc int = -1


type fnHead struct {
    name string
    args []arg
}

type arg struct {
    name string
    argType types.Type
}

func GetFn(funcName string) *fnHead {
    for _, f := range funcs {
        if f.name == funcName {
            return &f
        }
    }

    return nil
}


func Define(asm *os.File, op *prs.Op) {
    if op.Type != prs.OP_DEF_FN {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_DEF_FN\n")
        os.Exit(1)
    }

    if len(op.Operants) != 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_DEF_FN) should have 1 operant\n")
        os.Exit(1)
    }

    var f fnHead = fnHead{
        name: op.Operants[0],
    }
    curFunc = len(funcs)
    funcs = append(funcs, f)

    asm.WriteString(op.Operants[0] + ":\n")
}

func End(asm *os.File) {
    f := funcs[curFunc]

    // TODO: later local variables
    for _, a := range f.args {
        vars.Remove(a.name)
    }
    curFunc = -1

    asm.WriteString("ret\n\n")
}

func CallFunc(asm *os.File, op *prs.Op) {
    if op.Type != prs.OP_CALL_FN {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_CALL_FN\n")
        os.Exit(1)
    }

    if len(op.Operants) != 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_CALL_FN) should have 1 operant\n")
        os.Exit(1)
    }

    f := GetFn(op.Operants[0])
    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] undeclared name \"%s\"\n", op.Token.Str)
        fmt.Fprintln(os.Stderr, "\t" + op.Token.At())
        os.Exit(1)
    }

    asm.WriteString("call " + op.Operants[0] + "\n")
}

func DeclareArgs(op *prs.Op) {
    if op.Type != prs.OP_DEC_ARGS {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_DEC_ARGS\n")
        os.Exit(1)
    }

    if curFunc == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) CurFunc should be set")
        os.Exit(1)
    }

    f := &funcs[curFunc]

    name := ""
    for _, o := range op.Operants {
        if name == "" {
            name = o
        } else {
            if t := types.ToType(o); t != -1 {
                f.args = append(f.args, arg{ name: name, argType: t })
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] unknown type \"%s\"\n", o)
                os.Exit(1)
            }

            // see calling convention
            // 5 = r9, 6 = r10
            var regs []int
            var vartype types.Type

            const _ uint = 2 - types.TypesCount
            switch o {
            case "str":
                regs = []int { 5, 6 }
                vartype = types.Str

            case "i32":
                regs = []int { 5 }
                vartype = types.I32

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) unknown type \"%s\"", o)
                os.Exit(1)
            }

            vars.Add(vars.Var{Name: name, Regs: regs, Vartype: vartype})

            name = ""
        }
    }

    if len(f.args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only accept one argument max at the moment")
        os.Exit(1)
    }
}

func DefineArgs(asm *os.File, op *prs.Op) {
    if op.Type != prs.OP_DEF_ARGS {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_DEF_ARGS\n")
        os.Exit(1)
    }

    if len(op.Operants) < 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_DEF_ARGS) should have 1 or more operants\n")
        os.Exit(1)
    }

    if f := GetFn(op.Operants[0]); f != nil {
        for i, val := range op.Operants[1:] {
            if otherVar := vars.GetVar(val); otherVar != nil {
                if otherVar.Vartype != f.args[i].argType {
                    fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                        f.name, i, f.args[i].argType.Readable(), otherVar.Vartype.Readable())
                    fmt.Fprintln(os.Stderr, "\t" + op.Token.At())
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
            } else if prs.IsLit(val) {
                t := types.TypeOfVal(val)

                if t != f.args[i].argType {
                    fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                        f.name, i, f.args[i].argType.Readable(), t.Readable())
                    fmt.Fprintln(os.Stderr, "\t" + op.Token.At())
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
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", val)
                fmt.Fprintln(os.Stderr, "\t" + op.Token.At())
                os.Exit(1)
            }
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", op.Operants[0])
        os.Exit(1)
    }
}

func AddBuildIn(name string, argname string, argtype types.Type) {
    funcs = append(funcs, fnHead{name: name, args: []arg{{name: argname, argType: argtype}}})
}

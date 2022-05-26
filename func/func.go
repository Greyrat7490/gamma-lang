package fn

import (
    "os"
    "fmt"
    "strconv"
    "gorec/conditions"
    "gorec/loops"
    "gorec/str"
    "gorec/token"
    "gorec/types"
    "gorec/vars"
)

/*
System V AMD64 ABI calling convention
  * int   args1-6: rdi, rsi, rdx, rcx, r8, r9
  * float args1-8: xmm0 - xmm7

  * push args from right to left for more args
  * return value in eax/rax
  * caller cleans stack
  * callee reserves space (multiple of 16)
*/

var funcs []function
var curFunc int = -1

type function struct {
    Name token.Token
    Args []types.Type
}

func (f *function) At() string {
    return f.Name.At()
}

func GetCurFunc() *function {
    return &funcs[curFunc]
}

func GetFn(funcName string) *function {
    for _, f := range funcs {
        if f.Name.Str == funcName {
            return &f
        }
    }

    return nil
}

func Define(asm *os.File, fnName token.Token) {
    asm.WriteString(fnName.Str + ":\n")
    asm.WriteString("push rbp\nmov rbp, rsp\n")

    var f function = function{ Name: fnName }
    curFunc = len(funcs)
    funcs = append(funcs, f)
    vars.InGlobalScope = false
}

func ReserveSpace(asm *os.File, argsCount int, localVarsCount int) {
    size := (argsCount + localVarsCount) * 8
    if size > 0 {
        // size has to be the multiple of 16byte
        size += size % 16
        asm.WriteString(fmt.Sprintf("sub rsp, %d\n", size))
    }
}

func removeLocalVars(localVarsCount int) {
    count := localVarsCount + len(funcs[curFunc].Args)
    vars.RemoveLast(count)
}

func End(asm *os.File, localVarsCount int) {
    removeLocalVars(localVarsCount)

    vars.InGlobalScope = true
    curFunc = -1

    cond.ResetCount()
    loops.ResetCount()

    asm.WriteString("leave\n")
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

var regs []string = []string{ "rdi", "rsi", "rdx", "rcx", "r8", "r9" }
func DeclareArgs(asm *os.File, args []vars.Var) {
    if curFunc == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) CurFunc should be set")
        os.Exit(1)
    }

    f := &funcs[curFunc]
    for i, a := range args {
        f.Args = append(f.Args, a.Type)
        vars.Declare(a.Name, a.Type)

        asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", vars.GetLastVar().Offset, regs[i]))
        if a.Type == types.Str {
            asm.WriteString(fmt.Sprintf("mov QWORD [rbp-%d], %s\n", vars.GetLastVar().Offset+vars.VarSize, regs[i+1]))
        }
    }

    if len(f.Args) > 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] functions only accept one argument max at the moment")
        fmt.Fprintln(os.Stderr, "\t" + f.At())
        os.Exit(1)
    }
}

func DefineArgByValue(asm *os.File, fnName token.Token, argNum int, value token.Token) {
    if f := GetFn(fnName.Str); f != nil {
       if t := types.TypeOfVal(value.Str); t != -1 {
            if t != f.Args[argNum] {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                    f.Name.Str, argNum, f.Args[argNum].Readable(), t.Readable())
                os.Exit(1)
            }

            const _ uint = 3 - types.TypesCount
            switch t {
            case types.Str:
                strIdx := str.Add(value.Str)
                asm.WriteString(fmt.Sprintf("mov rdi, str%d\n", strIdx))
                asm.WriteString(fmt.Sprintf("mov rsi, %d\n", str.GetSize(strIdx)))

            case types.I32:
                i, _ := strconv.Atoi(value.Str)
                asm.WriteString(fmt.Sprintf("mov rdi, %d\n", i))

            case types.Bool:
                if value.Str == "true" {
                    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", 1))
                } else {
                    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", 0))
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
            if otherVar.Type != f.Args[argNum] {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%s\" but got \"%s\"\n",
                    f.Name.Str, argNum, f.Args[argNum].Readable(), otherVar.Type.Readable())
                fmt.Fprintln(os.Stderr, "\t" + fnName.At())
                os.Exit(1)
            }

            const _ uint = 3 - types.TypesCount
            switch otherVar.Type {
            case types.Str:
                s1, s2 := otherVar.Gets()
                asm.WriteString(fmt.Sprintf("mov rdi, %s\n", s1))
                asm.WriteString(fmt.Sprintf("mov rsi, %s\n", s2))

            case types.I32, types.Bool:
                asm.WriteString(fmt.Sprintf("mov rdi, %s\n", otherVar.Get()))

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) type of var \"%s\" is not correct\n", otherVar.Name.Str)
                os.Exit(1)
            }
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName.Str)
        os.Exit(1)
    }
}

func DefineArgByReg(asm *os.File, fnName token.Token, argNum int, reg string) {
    f := GetFn(fnName.Str)

    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName.Str)
        os.Exit(1)
    }

    if f.Args[argNum] == types.Str {
        fmt.Fprintf(os.Stderr, "[ERROR] expected function \"%s\" arg%d to be an i32 or bool but got str\n", fnName.Str, argNum)
        os.Exit(1)
    }

    asm.WriteString(fmt.Sprintf("mov rdi, %s\n", reg))
}

func AddBuildIn(name string, argname string, argtype types.Type) {
    funcs = append(funcs, function{
        Name: token.Token{ Str: name },
        Args: []types.Type{ argtype },
    })
}

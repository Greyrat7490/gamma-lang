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

var regs []vars.RegGroup = []vars.RegGroup{ vars.RegDi, vars.RegSi, vars.RegD, vars.RegC, vars.RegR8, vars.RegR9 }

var funcs []function
var curFunc int = -1

type function struct {
    Name token.Token
    Args []types.Type
}

func (f *function) At() string {
    return f.Name.At()
}

func GetFn(name string) *function {
    for _, f := range funcs {
        if f.Name.Str == name {
            return &f
        }
    }

    return nil
}

func Define(asm *os.File, name token.Token) {
    if name.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] function names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    asm.WriteString(name.Str + ":\n")
    asm.WriteString("push rbp\nmov rbp, rsp\n")

    var f function = function{ Name: name }
    curFunc = len(funcs)
    funcs = append(funcs, f)
}

func ReserveSpace(asm *os.File, argsSize int, blockSize int) {
    size := argsSize + blockSize
    if size > 0 {
        // size has to be the multiple of 16byte
        size += size % 16
        asm.WriteString(fmt.Sprintf("sub rsp, %d\n", size))
    }
}

func End(asm *os.File) {
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

func AddArg(argtype types.Type) {
    if curFunc == -1 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) curFunc should be set")
        os.Exit(1)
    }

    funcs[curFunc].Args = append(funcs[curFunc].Args, argtype)
}

func DefArg(asm *os.File, argnum int, argname token.Token, argtype types.Type) {
    if argtype.GetKind() == types.Str {
        asm.WriteString(fmt.Sprintf("mov %s [rbp-%d], %s\n", vars.GetWord(types.Ptr_Size), vars.GetLastOffset() + types.Ptr_Size,
            vars.GetReg(regs[argnum], types.Ptr_Size)))

        asm.WriteString(fmt.Sprintf("mov %s [rbp-%d], %s\n", vars.GetWord(types.I32_Size), vars.GetLastOffset(),
            vars.GetReg(regs[argnum+1], types.I32_Size)))
    } else {
        asm.WriteString(fmt.Sprintf("mov %s [rbp-%d], %s\n", vars.GetWord(argtype.Size()), vars.GetLastOffset(), vars.GetReg(regs[argnum], argtype.Size())))
    }
}


func PassVal(asm *os.File, fnName token.Token, argNum int, value token.Token) {
    if f := GetFn(fnName.Str); f != nil {
       if t := types.TypeOfVal(value.Str); t != nil {
            if t != f.Args[argNum] {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%v\" but got \"%v\"\n",
                    f.Name.Str, argNum, f.Args[argNum], t)
                os.Exit(1)
            }

            switch t.GetKind() {
            case types.Str:
                strIdx := str.Add(value)
                asm.WriteString(fmt.Sprintf("mov %s, _str%d\n", vars.GetReg(regs[argNum],   types.Ptr_Size), strIdx))
                asm.WriteString(fmt.Sprintf("mov %s, %d\n",     vars.GetReg(regs[argNum+1], types.I32_Size), str.GetSize(strIdx)))

            case types.I32:
                i, _ := strconv.Atoi(value.Str)
                asm.WriteString(fmt.Sprintf("mov %s, %d\n", vars.GetReg(regs[argNum], types.I32_Size), i))

            case types.Bool:
                if value.Str == "true" {
                    asm.WriteString(fmt.Sprintf("mov %s, %d\n", vars.GetReg(regs[argNum], types.Bool_Size), 1))
                } else {
                    asm.WriteString(fmt.Sprintf("mov %s, %d\n", vars.GetReg(regs[argNum], types.Bool_Size), 0))
                }

            case types.Ptr:
                fmt.Fprintln(os.Stderr, "TODO PtrType in PassVal")
                os.Exit(1)

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

func PassVar(asm *os.File, fnName token.Token, argNum int, varname token.Token) {
    if f := GetFn(fnName.Str); f != nil {
        if otherVar := vars.GetVar(varname.Str); otherVar != nil {
            if otherVar.GetType().GetKind() != f.Args[argNum].GetKind() {
                fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" takes as argument %d the type \"%v\" but got \"%v\"\n",
                    f.Name.Str, argNum, f.Args[argNum], otherVar.GetType())
                fmt.Fprintln(os.Stderr, "\t" + fnName.At())
                os.Exit(1)
            }

            switch otherVar.GetType().GetKind() {
            case types.Str:
                s1, s2 := otherVar.Gets()
                asm.WriteString(fmt.Sprintf("mov %s, %s\n", vars.GetReg(regs[argNum], types.Ptr_Size), s1))
                asm.WriteString(fmt.Sprintf("mov %s, %s\n", vars.GetReg(regs[argNum+1],   types.I32_Size), s2))

            case types.I32, types.Ptr:
                asm.WriteString(fmt.Sprintf("mov %s, %s\n", vars.GetReg(regs[argNum], otherVar.GetType().Size()), otherVar.Get()))

            case types.Bool:
                asm.WriteString(fmt.Sprintf("movzx %s, %s\n", vars.GetReg(regs[argNum], otherVar.GetType().Size()), otherVar.Get()))

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) type of var \"%s\" is not correct\n", varname.Str)
                os.Exit(1)
            }
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName.Str)
        os.Exit(1)
    }
}

func PassReg(asm *os.File, fnName token.Token, argNum int, reg vars.RegGroup) {
    f := GetFn(fnName.Str)

    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName.Str)
        os.Exit(1)
    }

    if f.Args[argNum].GetKind() == types.Str {
        fmt.Fprintf(os.Stderr, "[ERROR] expected function \"%s\" arg%d to be an i32, bool or a pointer but got str\n", fnName.Str, argNum)
        os.Exit(1)
    }

    size := f.Args[argNum].Size()
    if size < 2 {
        asm.WriteString(fmt.Sprintf("movzx %s, %s\n", vars.GetReg(regs[argNum], size), vars.GetReg(reg, size)))
    } else {
        asm.WriteString(fmt.Sprintf("mov %s, %s\n", vars.GetReg(regs[argNum], size), vars.GetReg(reg, size)))
    }
}


func AddBuildIn(name string, argname string, argtype types.Type) {
    funcs = append(funcs, function{
        Name: token.Token{ Str: name },
        Args: []types.Type{ argtype },
    })
}

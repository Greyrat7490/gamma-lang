package fn

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/loops"
    "gorec/token"
    "gorec/types"
    "gorec/types/str"
    "gorec/conditions"
    "gorec/asm/x86_64"
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

var regs []asm.RegGroup = []asm.RegGroup{ asm.RegDi, asm.RegSi, asm.RegD, asm.RegC, asm.RegR8, asm.RegR9 }

var funcs []function

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

func Declare(name token.Token, args []types.Type) {
    if name.Str[0] == '_' {
        fmt.Fprintln(os.Stderr, "[ERROR] function names starting with \"_\" are reserved for the compiler")
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    funcs = append(funcs, function{ Name: name, Args: args })
}

func Define(file *os.File, name token.Token, argsSize int, blockSize int) {
    f := GetFn(name.Str)
    if f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not declared\n", name.Str)
        fmt.Fprintln(os.Stderr, "\t" + name.At())
        os.Exit(1)
    }

    file.WriteString(name.Str + ":\n")
    file.WriteString("push rbp\nmov rbp, rsp\n")
    reserveSpace(file, argsSize, blockSize)
}

func reserveSpace(file *os.File, argsSize int, blockSize int) {
    size := argsSize + blockSize
    if size > 0 {
        // size has to be the multiple of 16byte
        size = (size + 15) & ^15
        file.WriteString(fmt.Sprintf("sub rsp, %d\n", size))
    }
}

func End(file *os.File) {
    cond.ResetCount()
    loops.ResetCount()

    file.WriteString("leave\n")
    file.WriteString("ret\n\n")
}

func CallFunc(file *os.File, fnName token.Token) {
    if f := GetFn(fnName.Str); f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] undeclared name \"%s\"\n", fnName.Str)
        fmt.Fprintln(os.Stderr, "\t" + fnName.At())
        os.Exit(1)
    }

    file.WriteString("call " + fnName.Str + "\n")
}

func DefArg(file *os.File, regIdx int, v *vars.LocalVar) {
    if v.Type.GetKind() == types.Str {
        setArg(file, v.Addr(0), regIdx, types.Ptr_Size)
        setArg(file, v.Addr(1), regIdx+1, types.I32_Size)
    } else {
        setArg(file, v.Addr(0), regIdx, v.Type.Size())
    }
}

func setArg(file *os.File, addr string, regIdx int, size int) {
    if regIdx >= len(regs) {
        fmt.Fprintf(os.Stderr, "[ERROR] not enough regs left to set args (max 6) %d more needed\n", regIdx - len(regs) + 1)
        os.Exit(1)
    }

    // adjust offset if reg is bigger than expected (no smaller reg available)
    if asm.GetSize(regs[regIdx], size) > size {
        file.WriteString(fmt.Sprintf("mov %s, %s\n", asm.GetReg(asm.RegA, asm.GetSize(regs[regIdx], size)), asm.GetReg(regs[regIdx], size)))
        file.WriteString(fmt.Sprintf("mov %s [%s], %s\n", asm.GetWord(size), addr, asm.GetReg(asm.RegA, size)))
    } else {
        file.WriteString(fmt.Sprintf("mov %s [%s], %s\n", asm.GetWord(size), addr, asm.GetReg(regs[regIdx], size)))
    }
}

func PassVal(file *os.File, fnName token.Token, regIdx int, value token.Token, valtype types.Type) {
    if f := GetFn(fnName.Str); f == nil {
        fmt.Fprintf(os.Stderr, "[ERROR] function \"%s\" is not defined", fnName.Str)
        os.Exit(1)
    }

    switch valtype.GetKind() {
    case types.Str:
        strIdx := str.Add(value)
        file.WriteString(asm.MovRegVal(regs[regIdx], types.Ptr_Size, fmt.Sprintf("_str%d", strIdx)))
        file.WriteString(asm.MovRegVal(regs[regIdx+1], types.I32_Size, fmt.Sprint(str.GetSize(strIdx))))

    case types.Bool:
        if value.Str == "true" { value.Str = "1" } else { value.Str = "0" }
        fallthrough

    case types.I32, types.Ptr:
        file.WriteString(asm.MovRegVal(regs[regIdx], valtype.Size(), value.Str))

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] could not get type of value \"%s\"\n", value.Str)
        os.Exit(1)
    }
}

func PassVar(file *os.File, regIdx int, otherVar vars.Var) {
    switch otherVar.GetType().GetKind() {
    case types.Str:
        file.WriteString(asm.MovRegDeref(regs[regIdx], otherVar.Addr(0), types.Ptr_Size))
        file.WriteString(asm.MovRegDeref(regs[regIdx+1], otherVar.Addr(1), types.I32_Size))

    case types.I32, types.Ptr:
        file.WriteString(asm.MovRegDeref(regs[regIdx], otherVar.Addr(0), otherVar.GetType().Size()))

    case types.Bool:
        size := otherVar.GetType().Size()
        // TODO: use movzx in asm funcs if needed
        file.WriteString(fmt.Sprintf("movzx %s, %s [%s]\n", asm.GetReg(regs[regIdx], size), asm.GetWord(size), otherVar.Addr(0)))

    default:
        // TODO
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) type of var is not correct\n")
        os.Exit(1)
    }
}

func PassReg(file *os.File, regIdx int, argType types.Type) {
    if argType.GetKind() == types.Str {
        file.WriteString(fmt.Sprintf("mov %s, %s\n", asm.GetReg(regs[regIdx], types.Ptr_Size), asm.GetReg(asm.RegA, types.Ptr_Size)))
        file.WriteString(fmt.Sprintf("mov %s, %s\n", asm.GetReg(regs[regIdx+1], types.I32_Size), asm.GetReg(asm.RegB, types.I32_Size)))
    } else {
        size := argType.Size()
        if size < 2 {
            file.WriteString(fmt.Sprintf("movzx %s, %s\n", asm.GetReg(regs[regIdx], size), asm.GetReg(asm.RegA, size)))
        } else {
            file.WriteString(fmt.Sprintf("mov %s, %s\n", asm.GetReg(regs[regIdx], size), asm.GetReg(asm.RegA, size)))
        }
    }
}


func AddBuildIn(name string, argname string, argtype types.Type) {
    funcs = append(funcs, function{
        Name: token.Token{ Str: name, Type: token.Name },
        Args: []types.Type{ argtype },
    })
}

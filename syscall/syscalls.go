package sys

import (
    "fmt"
    "os"
    "gorec/types"
    "gorec/func"
)

const STDOUT = 1

const SYS_WRITE = 1
const SYS_EXIT = 60

func syscall(asm *os.File, syscallNum uint, args... interface{}) {
    regs := []string{"rdi", "rsi", "rdx", "r10", "r8", "r9"}

    if len(args) > len(regs) {
        fmt.Fprintf(os.Stderr, "[ERROR] syscall only supports %d args\n", len(regs))
        os.Exit(1)
    }

    for i, arg := range args {
        asm.WriteString(fmt.Sprintf("mov %s, %s\n", regs[i], fmt.Sprint(arg)))
    }

    asm.WriteString(fmt.Sprintf("mov rax, %d\n", syscallNum))
    asm.WriteString("syscall\n")
}

func DefineWriteInt(asm *os.File) {
    args := []function.Arg{{Name: "i", ArgType: types.I32}}
    function.AddFunc(&function.Function{Name: "printInt", Args: args})

    asm.WriteString("printInt:\n")
    asm.WriteString("push rax\n")
    asm.WriteString("push rbx\n")
    asm.WriteString("push rcx\n")
    asm.WriteString("push rdx\n")
    asm.WriteString("mov rax, r9\n")
    asm.WriteString("call int_to_str\n")
    syscall(asm, SYS_WRITE, STDOUT, "rbx", "rax")
    asm.WriteString("pop rdx\n")
    asm.WriteString("pop rcx\n")
    asm.WriteString("pop rbx\n")
    asm.WriteString("pop rax\n")
    asm.WriteString("ret\n\n")
}

func DefineWriteStr(asm *os.File) {
    args := []function.Arg{{Name: "str", ArgType: types.Str}}
    function.AddFunc(&function.Function{Name: "printStr", Args: args})

    asm.WriteString("printStr:\n")
    asm.WriteString("push rax\n")
    asm.WriteString("push rcx\n")
    asm.WriteString("push rdx\n")
    syscall(asm, SYS_WRITE, STDOUT, "r9", "r10")
    asm.WriteString("pop rdx\n")
    asm.WriteString("pop rcx\n")
    asm.WriteString("pop rax\n")
    asm.WriteString("ret\n\n")
}

func DefineExit(asm *os.File) {
    args := []function.Arg{{Name: "i", ArgType: types.I32}}
    function.AddFunc(&function.Function{Name: "exit", Args: args})

    asm.WriteString("exit:\n")
    syscall(asm, SYS_EXIT, "r9")
    asm.WriteString("ret\n\n")
}

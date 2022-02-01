package main

import (
    "fmt"
    "os"
)

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

func defineWriteInt(asm *os.File) {
    args := []arg{{name: "i", argType: i32}}
    funcs = append(funcs, function{name: "printInt", args: args})

    asm.WriteString("printInt:\n")
    asm.WriteString("push rbx\n")
    asm.WriteString("push rax\n")
    asm.WriteString("mov rax, r9\n")
    asm.WriteString("call int_to_str\n")
    syscall(asm, SYS_WRITE, STDOUT, "rbx", "rax")
    asm.WriteString("pop rax\n")
    asm.WriteString("pop rbx\n")
    asm.WriteString("ret\n")
}

func defineWriteStr(asm *os.File) {
    args := []arg{{name: "str", argType: str}}
    funcs = append(funcs, function{name: "printStr", args: args})

    asm.WriteString("printStr:\n")
    syscall(asm, SYS_WRITE, STDOUT, "r9", "r10")
    asm.WriteString("ret\n")
}

func defineExit(asm *os.File) {
    args := []arg{{name: "i", argType: i32}}
    funcs = append(funcs, function{name: "exit", args: args})

    asm.WriteString("exit:\n")
    syscall(asm, SYS_EXIT, "r9")
    asm.WriteString("ret\n")
}

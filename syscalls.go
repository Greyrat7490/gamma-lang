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

// TODO: use stack to backup registers to prevent unwanted behavior
func write(asm *os.File, words []word, i int) int {
    args := getArgs(words[i:], 1)

    if args[0].isVar {
        v := vars[args[0].value]
        switch v.vartype {
        case str:
            if registers[v.regIdx].isAddr {
                syscall(asm, SYS_WRITE, STDOUT, registers[v.regIdx].name, strLits[registers[v.regIdx].value].size)
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] unreachable: register.isAddr should always be true if type of var is String")
                fmt.Fprintln(os.Stderr, "\t" + words[i].at())
                os.Exit(1)
            }

        case i32:
            if !registers[v.regIdx].isAddr {
                asm.WriteString("push rbx\n")
                asm.WriteString("push rax\n")
                asm.WriteString(fmt.Sprintf("mov rax, %s\n", registers[v.regIdx].name))
                asm.WriteString("call int_to_str\n")
                syscall(asm, SYS_WRITE, STDOUT, "rbx", "rax")
                asm.WriteString("pop rax\n")
                asm.WriteString("pop rbx\n")
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] unreachable: register.isAddr should always be false if type of var is Int")
                fmt.Fprintln(os.Stderr, "\t" + words[i].at())
                os.Exit(1)
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown type %d\n", v.vartype)
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
            os.Exit(1)
        }
    } else {
        switch args[0].argType {
        case str:
            syscall(asm, SYS_WRITE, STDOUT, fmt.Sprintf("str%d", args[0].value) , strLits[args[0].value].size)
        case i32:
            asm.WriteString("push rbx\n")
            asm.WriteString("push rax\n")
            asm.WriteString(fmt.Sprintf("mov rax, %d\n", args[0].value))
            asm.WriteString("call int_to_str\n")
            syscall(asm, SYS_WRITE, STDOUT, "rbx", "rax")
            asm.WriteString("pop rax\n")
            asm.WriteString("pop rbx\n")
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown type %d\n", args[0].argType)
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
            os.Exit(1)
        }
    }

    return i + len(args) + 2 // skip args, "(" and ")"
}

func exit(asm *os.File, words []word, i int) int {
    args := getArgs(words[i:], 1)

    if args[0].isVar {
        v := vars[args[0].value]
        switch v.vartype {
        case i32:
            syscall(asm, SYS_EXIT, registers[v.regIdx].name)
        case str:
            fmt.Fprintln(os.Stderr, "[ERROR] exit only accepts i32 (got str)")
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
            os.Exit(1)
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown type \"%d\"\n", args[0].argType)
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
            os.Exit(1)
        }
    } else {
        switch args[0].argType {
        case i32:
            syscall(asm, SYS_EXIT, args[0].value)
        case str:
            fmt.Fprintln(os.Stderr, "[ERROR] exit only accepts i32 (got str)")
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
            os.Exit(1)
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown type \"%d\"\n", args[0].argType)
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
            os.Exit(1)
        }
    }

    return i + len(args) + 2
}

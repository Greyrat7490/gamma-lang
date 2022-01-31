package main

import (
    "fmt"
    "os"
    "strconv"
)

const SYS_WRITE = 1
const SYS_EXIT = 60


type arg struct {
    isVar bool
    argType gType
    value int       // regIdx if isVar, strIdx if argType is str
}

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

func syscallParseArgs(words []word) (args []arg) {
    if len(words) < 1 || words[0].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[0].at())
        os.Exit(1)
    }

    b := false
    for _, w := range words[1:] {
        if w.str == ")" {
            b = true
            break
        }

        if w.str[0] == '"' && w.str[len(w.str) - 1] == '"' {
            args = append(args, arg{false, str, len(strLits)})
            addStrLit(w)
        } else if i, err := strconv.Atoi(w.str); err == nil {
            args = append(args, arg{false, i32, i})
        } else {
            if v := getVar(w.str); v != nil {
                args = append(args, arg{true, v.vartype, v.regIdx})
            } else {
                fmt.Fprintln(os.Stderr, vars)
                fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", w.str)
                fmt.Fprintln(os.Stderr, "\t" + w.at())
                os.Exit(1)
            }
        }
    }

    if !b {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    return args

}

// TODO: use stack to backup registers to prevent unwanted behavior
// TODO: to an actual function!!!!
func write(asm *os.File, words []word, i int) int {
    args := syscallParseArgs(words[i+1:])
    if len(args) != 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] function takes %d argument but got %d\n", 1, len(args))
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
        os.Exit(1)
    }

    if args[0].isVar {
        v := vars[args[0].value]
        switch v.vartype {
        case str:
            syscall(asm, SYS_WRITE, STDOUT, registers[v.regIdx].name, 5)

        case i32:
            asm.WriteString("push rbx\n")
            asm.WriteString("push rax\n")
            asm.WriteString(fmt.Sprintf("mov rax, %s\n", registers[v.regIdx].name))
            asm.WriteString("call int_to_str\n")
            syscall(asm, SYS_WRITE, STDOUT, "rbx", "rax")
            asm.WriteString("pop rax\n")
            asm.WriteString("pop rbx\n")

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
    args := syscallParseArgs(words[i+1:])
    if len(args) != 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] function takes %d argument but got %d\n", 1, len(args))
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
        os.Exit(1)
    }

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

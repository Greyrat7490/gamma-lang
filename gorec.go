package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "strings"
    "gorec/func"
    "gorec/parser"
    "gorec/syscall"
    "gorec/vars"
    "gorec/str"
)


func nasm_header(asm *os.File) {
    asm.WriteString("[BITS 64]\n")
    asm.WriteString("section .text\n")
    asm.WriteString("global _start\n")
}

func nasm_footer(asm *os.File) {
    asm.WriteString("\n_start:\n")
    asm.WriteString("mov rsp, stack_top\n")
    asm.WriteString("mov byte [intBuf + 11], 0xa\n\n")

    vars.WriteGlobalVars(asm)

    asm.WriteString("call main\n")

    asm.WriteString("\nmov rdi, 0\n")
    asm.WriteString(fmt.Sprintf("mov rax, %d\n", sys.SYS_EXIT))
    asm.WriteString("syscall\n")

    str.WriteStrLits(asm)

    asm.WriteString("\nsection .bss\n")
    asm.WriteString("\tresb 1024 * 1024\nstack_top:\n") // 1MiB
    asm.WriteString("intBuf:\n\tresb 12") // int(32bit) -> 10 digits max + \n and sign -> 12 char string max
}

func compile() {
    asm, err := os.Create("output.asm")
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] could not create \"output.asm\"")
        os.Exit(1)
    }
    defer asm.Close()

    nasm_header(asm)

    sys.DefineBuildIns(asm)

    for _, o := range prs.Ops {
        switch o.Type {
        case prs.OP_DEF_VAR:
            vars.Define(&o)
        case prs.OP_DEC_VAR:
            vars.Declare(&o)
        case prs.OP_CALL_FN:
            fn.CallFunc(asm, &o)
        case prs.OP_DEF_FN:
            fn.Define(asm, &o)
        case prs.OP_END_FN:
            fn.End(asm)
        case prs.OP_DEC_ARGS:
            fn.DeclareArgs(&o)
        case prs.OP_DEF_ARGS:
            fn.DefineArgs(asm, &o)
        default:
            fmt.Println(o.Type)
            fmt.Fprintln(os.Stderr, "TODO")
            os.Exit(1)
        }
    }

    nasm_footer(asm)
}

func genExe() {
    var stderr strings.Builder

    fmt.Println("[INFO] generating object files...")

    cmd := exec.Command("nasm", "-f", "elf64", "-o", "output.o", "output.asm")
    cmd.Stderr = &stderr
    err := cmd.Run()
    if err != nil {
        fmt.Println("[ERROR] ", stderr.String())
    }

    fmt.Println("[INFO] linking object files...")

    cmd = exec.Command("ld", "-o", "output", "output.o")
    cmd.Stderr = &stderr
    err = cmd.Run()
    if err != nil {
        fmt.Println("[ERROR] ", stderr.String())
    }

    fmt.Println("[INFO] generated executable")
}

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "[ERROR] you need to provide a source file to compile")
        os.Exit(1)
    }

    src, err := ioutil.ReadFile(os.Args[1])
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]", err)
        os.Exit(1)
    }

    // TODO: type checking step
    prs.Tokenize(src)
    compile()
    // TODO: optimization step

    genExe()
}

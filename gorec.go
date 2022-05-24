package main

import (
    "fmt"
    "os"
    "os/exec"
    "strings"
    "io/ioutil"
    "gorec/parser"
    "gorec/str"
    "gorec/syscall"
    "gorec/vars"
    "gorec/ast"
    "gorec/token"
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


    asm.WriteString("call main\n")

    asm.WriteString("\nmov rdi, 0\n")
    asm.WriteString(fmt.Sprintf("mov rax, %d\n", sys.SYS_EXIT))
    asm.WriteString("syscall\n")

    asm.WriteString("\nsection .data\n")
    vars.WriteGlobalVars(asm)
    str.WriteStrLits(asm)
    asm.WriteString("str_true: db \"true\", 0xa\n")
    asm.WriteString("str_false: db \"false\", 0xa\n")

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

    ast.Compile(asm)

    nasm_footer(asm)
}

func genExe() {
    var stderr strings.Builder

    fmt.Println("[INFO] generating object files...")

    cmd := exec.Command("nasm", "-f", "elf64", "-o", "output.o", "output.asm")
    cmd.Stderr = &stderr
    err := cmd.Run()
    if err != nil {
        s := stderr.String()
        if s == "" {
            fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] %v\n", s)
        }
        os.Exit(1)
    }

    fmt.Println("[INFO] linking object files...")

    cmd = exec.Command("ld", "-o", "output", "output.o")
    cmd.Stderr = &stderr
    err = cmd.Run()
    if err != nil {
        s := stderr.String()
        if s == "" {
            fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] %v\n", s)
        }
        os.Exit(1)
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

    token.Tokenize(src)
    prs.Parse()
    // TODO: type checking step
    // TODO: optimization step
    ast.ShowAst()
    compile()
    genExe()
}

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
    "gorec/parser"
    "gorec/vars"
    "gorec/func"
    "gorec/syscall"
)


func nasm_header(asm *os.File) {
    asm.WriteString("[BITS 64]\n")
    asm.WriteString("section .text\n")
    asm.WriteString("global _start\n")

    asm.WriteString(`; rax = input int
; rbx = output string pointer
; rax = output string length
uint_to_str:
    push rcx
    push rdx

    mov ecx, 10

    mov rbx, intBuf + 10
    .l1:
        xor edx, edx
        div ecx
        add dl, 48
        dec rbx
        mov byte [rbx], dl
        cmp eax, 0
        jne .l1

    mov rax, rbx
    sub rax, intBuf
    inc rax
    pop rdx
    pop rcx
    ret

int_to_str:
    push rcx
    push rdx
    push rax

    mov ecx, 10
    mov rbx, intBuf + 10

    cmp rax, 0
    jge .l1

    neg rax

    .l1:
        xor edx, edx
        div ecx
        add dl, 48
        dec rbx
        mov byte [rbx], dl
        cmp eax, 0
        jne .l1

    pop rax
    cmp rax, 0
    jge .end

    dec rbx
    mov byte [rbx], 0x2d

    .end:
        mov rax, rbx
        sub rax, intBuf
        inc rax
        pop rdx
        pop rcx
        ret

`)
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

    vars.WriteStrLits(asm)

    asm.WriteString("\nsection .bss\n")
    asm.WriteString("\tresb 1024 * 1024\nstack_top:\n") // 1MiB
    asm.WriteString("intBuf:\n\tresb 12") // int(32bit) -> 10 digits max + \n and sign -> 12 char string max
}

func compile(srcFile []byte) {
    asm, err := os.Create("output.asm")
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] could not create \"output.asm\"")
        os.Exit(1)
    }
    defer asm.Close()

    nasm_header(asm)

    // define build-in functions
    // TODO: add int_to_str and uint_to_str
    sys.DefineWriteStr(asm)
    sys.DefineWriteInt(asm)
    sys.DefineExit(asm)

    prs.Split(string(srcFile))

    for i := 0; i < len(prs.Words); i++ {
        switch prs.Words[i].Str {
        case "var":
            i = vars.DeclareVar(prs.Words, i)
        case ":=":
            i = vars.DefineVar(prs.Words, i)
        case "fn":
            i = function.DefineFunc(asm, prs.Words, i)
        case "printInt":
            fmt.Fprintln(os.Stderr, "[ERROR] function calls outside of main are not allowed")
            fmt.Fprintln(os.Stderr, "\t" + prs.Words[i].At())
            os.Exit(1)
        case "printStr":
            fmt.Fprintln(os.Stderr, "[ERROR] function calls outside of main are not allowed")
            fmt.Fprintln(os.Stderr, "\t" + prs.Words[i].At())
            os.Exit(1)
        case "exit":
            fmt.Fprintln(os.Stderr, "[ERROR] function calls outside of main are not allowed")
            fmt.Fprintln(os.Stderr, "\t" + prs.Words[i].At())
            os.Exit(1)
        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown word \"%s\"\n", prs.Words[i].Str)
            fmt.Fprintln(os.Stderr, "\t" + prs.Words[i].At())
            os.Exit(1)
        }
    }

    function.Checks();

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
    compile(src)
    // TODO: optimization step

    genExe()
}

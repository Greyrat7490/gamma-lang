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

func DefineBuildIns(asm *os.File) {
    defineItoS(asm)
    defineBtoS(asm)
    defineWriteStr(asm)
    defineWriteInt(asm)
    defineWriteBool(asm)
    defineExit(asm)
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

    asm.WriteString("push rcx\n")
    asm.WriteString("push r11\n")   // syscall can change r11 and rcx

    asm.WriteString("syscall\n")

    asm.WriteString("pop r11\n")
    asm.WriteString("pop rcx\n")
}

func defineWriteInt(asm *os.File) {
    fn.AddBuildIn("printInt", "i", types.I32)

    asm.WriteString("printInt:\n")
    asm.WriteString("push rax\n")
    asm.WriteString("push rbx\n")
    asm.WriteString("push rdx\n")

    asm.WriteString("mov rax, r10\n")
    asm.WriteString("call int_to_str\n")
    syscall(asm, SYS_WRITE, STDOUT, "rbx", "rax")

    asm.WriteString("pop rdx\n")
    asm.WriteString("pop rbx\n")
    asm.WriteString("pop rax\n")
    asm.WriteString("ret\n\n")
}

func defineWriteBool(asm *os.File) {
    fn.AddBuildIn("printBool", "b", types.Bool)

    asm.WriteString("printBool:\n")
    asm.WriteString("push rax\n")
    asm.WriteString("push rbx\n")
    asm.WriteString("push rdx\n")

    asm.WriteString("mov rax, r10\n")
    asm.WriteString("call bool_to_str\n")
    syscall(asm, SYS_WRITE, STDOUT, "rbx", "rax")

    asm.WriteString("pop rdx\n")
    asm.WriteString("pop rbx\n")
    asm.WriteString("pop rax\n")
    asm.WriteString("ret\n\n")
}

func defineWriteStr(asm *os.File) {
    fn.AddBuildIn("printStr", "s", types.Str)

    asm.WriteString("printStr:\n")
    asm.WriteString("push rax\n")
    asm.WriteString("push rdx\n")

    syscall(asm, SYS_WRITE, STDOUT, "r10", "r11")

    asm.WriteString("pop rdx\n")
    asm.WriteString("pop rax\n")
    asm.WriteString("ret\n\n")
}

func defineExit(asm *os.File) {
    fn.AddBuildIn("exit", "i", types.I32)

    asm.WriteString("exit:\n")
    syscall(asm, SYS_EXIT, "r9")
    asm.WriteString("ret\n\n")
}

func defineBtoS(asm *os.File) {
    asm.WriteString(`; rax = input int
; rbx = output string pointer
; rax = output string length
bool_to_str:
    cmp rax, 0
    jne .c1
    mov rbx, str_false
    mov rax, 6
    ret
    .c1:
        mov rbx, str_true
        mov rax, 5
        ret

`)
}

func defineItoS(asm *os.File) {
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

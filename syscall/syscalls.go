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

// linux syscall calling convention
// arg: 0    1    2    3   4   5
//     rdi, rsi, rdx, r10, r8, r9
// return: rax
func syscall(asm *os.File, syscallNum uint) {
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
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func defineWriteBool(asm *os.File) {
    fn.AddBuildIn("printBool", "b", types.Bool)

    asm.WriteString("printBool:\n")

    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call bool_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func defineWriteStr(asm *os.File) {
    fn.AddBuildIn("printStr", "s", types.Str)

    asm.WriteString("printStr:\n")

    asm.WriteString("mov rdx, rsi\n")
    asm.WriteString("mov rsi, rdi\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func defineExit(asm *os.File) {
    fn.AddBuildIn("exit", "i", types.I32)

    asm.WriteString("exit:\n")
    asm.WriteString(fmt.Sprintf("mov rax, %d\n", SYS_EXIT))
    asm.WriteString("syscall\n")
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

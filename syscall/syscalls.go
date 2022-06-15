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
    definePrintStr(asm)
    definePrintInt(asm)
    definePrintPtr(asm)
    definePrintBool(asm)
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

func definePrintInt(asm *os.File) {
    fn.AddBuildIn("printInt", "i", types.I32Type{})

    asm.WriteString("printInt:\n")
    asm.WriteString("movsxd rax, edi\n")   // mov edi into eax and sign extends upper half of rax
    asm.WriteString("call _int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func definePrintPtr(asm *os.File) {
    fn.AddBuildIn("printPtr", "i", types.PtrType{})

    asm.WriteString("printPtr:\n")
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func definePrintBool(asm *os.File) {
    fn.AddBuildIn("printBool", "b", types.BoolType{})

    asm.WriteString("printBool:\n")

    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _bool_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func definePrintStr(asm *os.File) {
    fn.AddBuildIn("printStr", "s", types.StrType{})

    asm.WriteString("printStr:\n")

    asm.WriteString("mov rdx, rsi\n")
    asm.WriteString("mov esi, edi\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func defineExit(asm *os.File) {
    fn.AddBuildIn("exit", "i", types.I32Type{})

    asm.WriteString("exit:\n")
    asm.WriteString(fmt.Sprintf("mov rax, %d\n", SYS_EXIT))
    asm.WriteString("syscall\n")
    asm.WriteString("ret\n\n")
}

func defineBtoS(asm *os.File) {
    asm.WriteString(`; rax = input int
; rbx = output string pointer
; rax = output string length
_bool_to_str:
    cmp rax, 0
    jne .c1
    mov rbx, _false
    mov rax, 5
    ret
    .c1:
        mov rbx, _true
        mov rax, 4
        ret

`)
}

func defineItoS(asm *os.File) {
    asm.WriteString(`; rax = input int
; rbx = output string pointer
; rax = output string length
_uint_to_str:
    push rcx
    push rdx

    mov ecx, 10

    mov rbx, _intBuf + 20
    .l1:
        xor edx, edx
        div ecx
        add dl, 48
        dec rbx
        mov byte [rbx], dl
        cmp eax, 0
        jne .l1

    mov rax, rbx
    sub rax, _intBuf
    inc rax
    pop rdx
    pop rcx
    ret

_int_to_str:
    push rcx
    push rdx
    push rax

    mov ecx, 10
    mov rbx, _intBuf + 20

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
        sub rax, _intBuf
        inc rax
        pop rdx
        pop rcx
        ret

`)
}

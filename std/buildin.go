package std

import (
    "os"
    "fmt"
    "gamma/types"
    "gamma/ast/identObj"
)

const STDOUT = 1

const SYS_WRITE = 1
const SYS_EXIT = 60


func Declare() {
    // build-in funcs
    identObj.AddBuildIn("printStr",  types.CreateStr(), nil)
    identObj.AddBuildIn("printInt",  types.CreateInt(types.I64_Size), nil)
    identObj.AddBuildIn("printUint", types.CreateUint(types.U64_Size), nil)
    identObj.AddBuildIn("printPtr",  types.PtrType{}, nil)
    identObj.AddBuildIn("printBool", types.BoolType{}, nil)
    identObj.AddBuildIn("printChar", types.CharType{}, nil)
    identObj.AddBuildIn("exit",      types.CreateInt(types.I32_Size), nil)

    // "inline assembly"
    identObj.AddBuildIn("_syscall", types.CreateInt(types.I64_Size), nil)
}

func Define(file *os.File) {
    defineItoS(file)
    defineBtoS(file)

    definePrintStr(file)
    definePrintChar(file)
    definePrintInt(file)
    definePrintUint(file)
    definePrintPtr(file)
    definePrintBool(file)

    defineExit(file)
    file.WriteString("\n")
}


// linux syscall calling convention like System V AMD64 ABI
func syscall(file *os.File, syscallNum uint) {
    file.WriteString(fmt.Sprintf("mov rax, %d\n", syscallNum))
    file.WriteString("syscall\n")
}

func definePrintStr(asm *os.File) {
    asm.WriteString("printStr:\n")

    asm.WriteString("mov rdx, rsi\n")
    asm.WriteString("mov esi, edi\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintChar(asm *os.File) {
    asm.WriteString("printChar:\n")
    asm.WriteString("mov ax, di\n")
    asm.WriteString("mov byte [_intBuf], al\n")
    asm.WriteString("mov rdx, 1\n")
    asm.WriteString("mov rsi, _intBuf\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintInt(asm *os.File) {
    asm.WriteString("printInt:\n")
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintUint(asm *os.File) {
    asm.WriteString("printUint:\n")
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _uint_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintPtr(asm *os.File) {
    asm.WriteString("printPtr:\n")
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintBool(asm *os.File) {
    asm.WriteString("printBool:\n")

    asm.WriteString("mov ax, di\n")
    asm.WriteString("call _bool_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func defineBtoS(asm *os.File) {
    asm.WriteString(
`_bool_to_str:
    cmp ax, 0
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
    asm.WriteString(
`; rax = input int
; rbx = output string pointer
; rax = output string length
_uint_to_str:
    mov rcx, 10
    lea rbx, [_intBuf+21]
    .l1:
        xor rdx, rdx
        div rcx
        add dl, 48
        dec rbx
        mov byte [rbx], dl
        cmp rax, 0
        jne .l1
    lea rax, [_intBuf+21]
    sub rax, rbx
    ret
_int_to_str:
    cmp rax, 0
    jge _uint_to_str
    neg rax
    mov rcx, 10
    lea rbx, [_intBuf+21]
    .l1:
        xor rdx, rdx
        div rcx
        add dl, 48
        dec rbx
        mov byte [rbx], dl
        cmp rax, 0
        jne .l1
    dec rbx
    mov byte [rbx], 0x2d
    lea rax, [_intBuf+21]
    sub rax, rbx
    ret
`)
}

func defineExit(file *os.File) {
    file.WriteString("exit:\n")
    syscall(file, SYS_EXIT)
}

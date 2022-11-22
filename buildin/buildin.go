package buildin

import (
    "fmt"
    "bufio"
    "gamma/types"
    "gamma/ast/identObj"
)

const SYS_WRITE = 1
const SYS_MMAP  = 9
const SYS_EXIT  = 60

const STDOUT = 1

const PROT_READ     = 1
const PROT_WRITE    = 2
const MAP_ANONYMOUS = 0x20
const MAP_PRIVATE   = 2


func Declare() {
    // build-in funcs
    identObj.AddBuildIn("printStr",  types.CreateStr(), nil)
    identObj.AddBuildIn("printInt",  types.CreateInt(types.I64_Size), nil)
    identObj.AddBuildIn("printUint", types.CreateUint(types.U64_Size), nil)
    identObj.AddBuildIn("printPtr",  types.PtrType{}, nil)
    identObj.AddBuildIn("printBool", types.BoolType{}, nil)
    identObj.AddBuildIn("printChar", types.CharType{}, nil)
    identObj.AddBuildIn("exit",      types.CreateInt(types.I32_Size), nil)

    // basic inline assembly
    identObj.AddBuildIn("_asm", types.CreateStr(), nil)
    identObj.AddBuildIn("_syscall", types.CreateInt(types.I64_Size), types.CreateInt(types.I64_Size))
}

func Define(file *bufio.Writer) {
    defineItoS(file)
    defineBtoS(file)

    definePrintStr(file)
    definePrintChar(file)
    definePrintInt(file)
    definePrintUint(file)
    definePrintPtr(file)
    definePrintBool(file)

    defineAppend(file)
    defineAllocVec(file)
    
    defineExit(file)
    file.WriteString("\n")
}


// linux syscall calling convention like System V AMD64 ABI
func syscall(file *bufio.Writer, syscallNum uint) {
    file.WriteString(fmt.Sprintf("mov rax, %d\n", syscallNum))
    file.WriteString("syscall\n")
}

func definePrintStr(asm *bufio.Writer) {
    asm.WriteString("printStr:\n")

    asm.WriteString("mov rdx, rsi\n")
    asm.WriteString("mov esi, edi\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintChar(asm *bufio.Writer) {
    asm.WriteString("printChar:\n")
    asm.WriteString("mov byte [_intBuf], dil\n")
    asm.WriteString("mov rdx, 1\n")
    asm.WriteString("mov rsi, _intBuf\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintInt(asm *bufio.Writer) {
    asm.WriteString("printInt:\n")
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintUint(asm *bufio.Writer) {
    asm.WriteString("printUint:\n")
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _uint_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintPtr(asm *bufio.Writer) {
    asm.WriteString("printPtr:\n")
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintBool(asm *bufio.Writer) {
    asm.WriteString("printBool:\n")

    asm.WriteString("mov eax, edi\n")
    asm.WriteString("call _bool_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func defineBtoS(asm *bufio.Writer) {
    asm.WriteString(
`_bool_to_str:
    test eax, eax
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

func defineItoS(asm *bufio.Writer) {
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
        add edx, 48
        dec rbx
        mov byte [rbx], dl
        test rax, rax
        jne .l1
    lea rax, [_intBuf+21]
    sub rax, rbx
    ret
_int_to_str:
    test rax, rax
    jge _uint_to_str
    neg rax
    mov rcx, 10
    lea rbx, [_intBuf+21]
    .l1:
        xor rdx, rdx
        div rcx
        add edx, 48
        dec rbx
        mov byte [rbx], dl
        test rax, rax
        jne .l1
    dec rbx
    mov byte [rbx], 0x2d
    lea rax, [_intBuf+21]
    sub rax, rbx
    ret
`)
}

func defineExit(file *bufio.Writer) {
    file.WriteString("exit:\n")
    syscall(file, SYS_EXIT)
}

func defineAppend(file *bufio.Writer) {
    
}

// TODO: rather use std/memory malloc
func defineAllocVec(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf(`
; rax = input size
; rax = output pointer
_alloc_vec:
mov rdi, 0
mov rsi, rax
mov rdx, %d
mov r10, %d
mov r8, -1
mov r9, 0
`, PROT_READ | PROT_WRITE, MAP_ANONYMOUS | MAP_PRIVATE))
    syscall(file, SYS_MMAP)
    file.WriteString("ret\n")
}

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
const STDERR = 2

const PROT_READ     = 1
const PROT_WRITE    = 2
const MAP_ANONYMOUS = 0x20
const MAP_PRIVATE   = 2


func Declare() {
    // build-in funcs
    identObj.AddBuildIn("print", types.StrType{}, nil)
    identObj.AddBuildIn("println", types.StrType{}, nil)
    identObj.AddBuildIn("eprint", types.StrType{}, nil)
    identObj.AddBuildIn("eprintln", types.StrType{}, nil)

    identObj.AddBuildIn("exit", types.CreateInt(types.I32_Size), nil)

    identObj.AddBuildIn("itos", types.CreateInt(types.I64_Size), types.StrType{})
    identObj.AddBuildIn("utos", types.CreateUint(types.U64_Size), types.StrType{})
    identObj.AddBuildIn("btos", types.BoolType{}, types.StrType{})
    identObj.AddBuildIn("ctos", types.CharType{}, types.StrType{})
    identObj.AddBuildIn("from_cstr", types.PtrType{ BaseType: types.CharType{} }, types.StrType{})

    identObj.AddBuildIn("fmt", nil, types.StrType{})
    identObj.AddGenBuildIn("sizeof", "T", nil, types.CreateUint(types.Ptr_Size))

    // basic inline assembly
    identObj.AddBuildIn("_asm", types.StrType{}, nil)
    identObj.AddBuildIn("_syscall", types.CreateInt(types.I64_Size), types.CreateInt(types.I64_Size))
}

func Define(file *bufio.Writer) {
    definePrint(file)
    definePrintln(file)
    defineEprint(file)
    defineEprintln(file)

    defineStrCmp(file)
    defineStrConcat(file)

    defineFromCStr(file)
    defineItoS(file)
    defineBtoS(file)
    defineCtoS(file)

    defineExit(file)
    file.WriteString("\n")
}


// linux syscall calling convention like System V AMD64 ABI
func syscall(file *bufio.Writer, syscallNum uint) {
    file.WriteString(fmt.Sprintf("mov rax, %d\n", syscallNum))
    file.WriteString("syscall\n")
}

func definePrint(asm *bufio.Writer) {
    asm.WriteString("print:\n")

    asm.WriteString("mov edx, esi\n")
    asm.WriteString("mov rsi, rdi\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func defineEprint(asm *bufio.Writer) {
    asm.WriteString("eprint:\n")

    asm.WriteString("mov edx, esi\n")
    asm.WriteString("mov rsi, rdi\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDERR))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n")
}

func definePrintln(asm *bufio.Writer) {
    asm.WriteString("println:\n")
    asm.WriteString("call print\n")
    asm.WriteString(fmt.Sprintf(
`mov rdi, 1
call malloc
mov byte [rax], %d
mov rdi, rax
mov esi, 1
`, int('\n')))
    asm.WriteString("call print\n")
    asm.WriteString("ret\n")
}

func defineEprintln(asm *bufio.Writer) {
    asm.WriteString("eprintln:\n")
    asm.WriteString("call eprint\n")
    asm.WriteString(fmt.Sprintf(
`mov rdi, 1
call malloc
mov byte [rax], %d
mov rdi, rax
mov esi, 1
`, int('\n')))
    asm.WriteString("call eprint\n")
    asm.WriteString("ret\n")
}

func defineCtoS(asm *bufio.Writer) {
    asm.WriteString(
`ctos:
    mov rbx, rdi
    mov rdi, 1
    call malloc
    mov byte [rax], bl
    mov edx, 1
    ret
`)
}

// TODO malloc and copy
func defineBtoS(asm *bufio.Writer) {
    asm.WriteString(
`btos:
    test edi, edi
    jne .c1
    mov rax, _false
    mov edx, 5
    ret
    .c1:
        mov rax, _true
        mov edx, 4
        ret
`)
}

func defineItoS(asm *bufio.Writer) {
 // max 64bit -> 20 digits max + sign -> 21 char string max
    asm.WriteString(
`utos:
    mov rbx, rdi
    mov rdi, 21
    call malloc
    lea rsi, [rax+21]
    mov rax, rbx
    mov rbx, rsi
    mov rcx, 10
    .l1:
        xor rdx, rdx
        div rcx
        add edx, 48
        dec rbx
        mov byte [rbx], dl
        test rax, rax
        jne .l1
    mov rdx, rsi
    sub rdx, rbx
    mov rax, rbx
    ret
itos:
    test rdi, rdi
    jge utos
    mov rbx, rdi
    mov rdi, 21
    call malloc
    lea rsi, [rax+21]
    mov rax, rbx
    mov rbx, rsi
    neg rax
    mov rcx, 10
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
    mov rdx, rsi
    sub rdx, rbx
    mov rax, rbx
    ret
`)
}

func defineExit(file *bufio.Writer) {
    file.WriteString("exit:\n")
    syscall(file, SYS_EXIT)
}

func defineFromCStr(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf(`from_cstr:
lea rdx, [rdi-1]
.l1:
inc rdx
cmp BYTE [rdx], 0
jne .l1
inc rdx
sub rdx, rdi
mov rax, rdi
ret
`))
}

func defineStrCmp(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf(`
; rax = ptr1
; edx = size1
; rbx = ptr2
; ecx = size2
; eax = output bool
_str_cmp:
cmp edx, ecx
jne .unequ
.l1:
movzx ecx, BYTE [rax]
cmp BYTE [rbx], cl
jne .unequ
inc rax
inc rbx
dec edx
cmp edx, 0
jg .l1
mov eax, 1
ret
.unequ:
mov eax, 0
ret
`))
}

func defineStrConcat(file *bufio.Writer) {
    file.WriteString(fmt.Sprintf(`
; rax = ptr1
; edx = size1
; rbx = ptr2
; ecx = size2
; rax = res ptr
; edx = res size
_str_concat:
push rbp
mov rbp, rsp
sub rsp, 32
mov QWORD [rbp-8], rax
mov QWORD [rbp-16], rbx
mov DWORD [rbp-20], ecx
mov DWORD [rbp-24], edx

mov edi, edx
add edi, ecx
mov DWORD [rbp-28], edi

call malloc

mov rbx, QWORD [rbp-8]
mov edx, DWORD [rbp-24]
.l1:
movzx ecx, BYTE [rbx]
mov BYTE [rax], cl
inc rax
inc rbx
dec edx
cmp edx, 0
jg .l1

mov ecx, DWORD [rbp-20]
mov rbx, QWORD [rbp-16]
.l2:
movzx edx, BYTE [rbx]
mov BYTE [rax], dl
inc rax
inc rbx
dec ecx
cmp ecx, 0
jg .l2

mov edx, DWORD [rbp-28]
sub rax, rdx

leave
ret
`))
}

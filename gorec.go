package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "unicode"
)

const SYS_WRITE = 1
const SYS_EXIT = 60

const STDOUT = 1

type reg struct {
    name string
    isAddr bool
    value int      // either an actual value or an address(index)
}

type vType int
const (
    Int vType = iota
    String vType = iota
)

type arg struct {
    isVar bool
    regIdx int
}

// TODO: proper string literales ("" to indicate)
var strLits []string

// TODO: register allocator for variables
var registers []reg = []reg { // so far safe to use registers for variables
    {name: "rbx"},
    {name: "r9"},
    {name: "r10"},
}


func nasm_header(asm *os.File) {
    asm.WriteString("[BITS 64]\n")
    asm.WriteString("section .text\n")
    asm.WriteString("global _start\n")

    asm.WriteString(`; rax = input int
; rbx = output string pointer
; rax = output string length
int_to_str:
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
    pop rdx
    pop rcx
    ret
`)

    asm.WriteString("\n_start:\n")
    asm.WriteString("mov rsp, stack_top\n\n")
}

func nasm_footer(asm *os.File) {
    asm.WriteString("\nmov rdi, 0\n")
    asm.WriteString(fmt.Sprintf("mov rax, %d\n", SYS_EXIT))
    asm.WriteString("syscall\n")
    asm.WriteString("\nsection .data\n")

    if len(strLits) > 0 {
        for i, str := range strLits {
            asm.WriteString(fmt.Sprintf("str%d: db \"%s\", 0xa\n", i, str))
        }
    }

    asm.WriteString("\nsection .bss\n")
    asm.WriteString("\tresb 1024 * 1024\nstack_top:\n") // 1MiB
    asm.WriteString("intBuf:\n\tresb 10") // int(32bit) -> 10 digits max -> 10 char string max
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

func getArgs(words []string, expectedArgCount int) (args []arg) {
    if len(words) < 2 || words[1] != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        os.Exit(1)
    }

    for _, w := range words[2:] {
        if w == ")" {
            break
        }

        if v := getVar(w); v != nil {           // variable
            args = append(args, arg{true, v.regIdx})
        } else {                                // string/int literal
            args = append(args, arg{false, len(strLits)})
            strLits = append(strLits, w)
        }
    }

    if len(words) - 2 == len(args) {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    if len(args) != expectedArgCount {
        fmt.Fprintf(os.Stderr, "[ERROR] function takes %d argument but got %d\n", expectedArgCount, len(args))
        os.Exit(1)
    }

    return args
}


// TODO: use stack to backup registers to prevent unwanted behavior
func write(asm *os.File, words []string, i int) int {
    args := getArgs(words[i:], 1)

    if args[0].isVar {
        v := vars[args[0].regIdx]
        switch v.vartype {
        case String:
            if registers[v.regIdx].isAddr {
                syscall(asm, SYS_WRITE, STDOUT, registers[v.regIdx].name, len(strLits[registers[v.regIdx].value]) + 1)
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] unreachable: register.isAddr should always be true if type of var is String")
                os.Exit(1)
            }

        // TODO: add linebreak
        // TODO: add sign
        case Int:
            if !registers[v.regIdx].isAddr {
                asm.WriteString("push rbx\n")
                asm.WriteString("push rax\n")
                asm.WriteString(fmt.Sprintf("mov rax, %s\n", registers[v.regIdx].name))
                asm.WriteString("call int_to_str\n")
                syscall(asm, SYS_WRITE, STDOUT, "rbx", "rax")
                asm.WriteString("pop rax\n")
                asm.WriteString("pop rbx\n")
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] unreachable: register.isAddr should always be false if type of var is Int")
                os.Exit(1)
            }

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] unknown type \"%#v\"\n", v.vartype)
            os.Exit(1)
        }
    } else {
        syscall(asm, SYS_WRITE, STDOUT, fmt.Sprintf("str%d", args[0].regIdx) , len(strLits[args[0].regIdx]) + 1)
    }

    return i + len(args) + 2 // skip args, "(" and ")"
}

// TODO: multiline comment
func split(file string) (words []string) {
    start := 0

    skip := false
    mlSkip := false

    for i, r := range(file) {
        if skip {
            if mlSkip {
                if r == '*' && file[i+1] == '/' {
                    skip = false
                    mlSkip = false
                    start = i + 2
                }
            } else {
                if r == '\n' {
                    skip = false
                    start = i + 1
                }

            }
            continue
        }

        if r == '/' {
            if file[i+1] == '/' {
                skip = true
            } else if file[i+1] == '*' {
                skip = true
                mlSkip = true
            }
        }

        if unicode.IsSpace(r) || r == '(' || r == ')' {
            if start != i {
                words = append(words, file[start:i])
            }
            start = i + 1

            if r == '(' || r == ')' {
                words = append(words, string(r))
            }
        }
    }

    if mlSkip {
        fmt.Fprintln(os.Stderr, "you have not terminated your comment (missing \"*/\")")
        os.Exit(1)
    }

    return words
}

func compile(srcFile []byte) {
    asm, err := os.Create("output.asm")
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] could not create \"output.asm\"")
        os.Exit(1)
    }
    defer asm.Close()

    nasm_header(asm)

    words := split(string(srcFile))

    for i := 0; i < len(words); i++ {
        switch words[i] {
        case "println":
            i = write(asm, words, i)
        case "var":
            i = declareVar(words, i)
        case ":=":
            i = defineVar(asm, words, i)

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] keyword \"%s\" is not supported\n", words[i])
            os.Exit(1)
        }
    }

    nasm_footer(asm)
}

func genExe() {
    cmd := exec.Command("nasm", "-f", "elf64", "-o", "output.o", "output.asm")
    err := cmd.Run()
    // TODO: better error messages
    checkErr(err)

    cmd = exec.Command("ld", "-o", "output", "output.o")
    err = cmd.Run()
    checkErr(err)
}

func checkErr(err error) {
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]", err)
        os.Exit(1)
    }
}

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "[ERROR] you need to provide a source file to compile")
        os.Exit(1)
    }

    src, err := ioutil.ReadFile(os.Args[1])
    checkErr(err)

    // TODO: type checking step
    compile(src)
    // TODO: optimization step

    genExe()
}

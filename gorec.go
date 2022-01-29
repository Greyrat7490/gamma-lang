package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "unicode"
    "strings"
    "strconv"
)

const STDOUT = 1

type arg struct {
    isVar bool
    argType gType
    value int       // regIdx if isVar, strIdx if argType is str
}

type word struct {
    line int
    col int
    str string
    // later filename
}

func (w word) at() string {
    return fmt.Sprintf("at line: %d, col: %d", w.line, w.col)
}


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

    asm.WriteString("\n_start:\n")
    asm.WriteString("mov rsp, stack_top\n")
    asm.WriteString("mov byte [intBuf + 11], 0xa\n\n")
}

func nasm_footer(asm *os.File) {
    asm.WriteString("\nmov rdi, 0\n")
    asm.WriteString(fmt.Sprintf("mov rax, %d\n", SYS_EXIT))
    asm.WriteString("syscall\n")

    if len(strLits) > 0 {
        asm.WriteString("\nsection .data\n")

        for i, str := range strLits {
            asm.WriteString(fmt.Sprintf("str%d: db %s\n", i, str.value))
        }
    }

    asm.WriteString("\nsection .bss\n")
    asm.WriteString("\tresb 1024 * 1024\nstack_top:\n") // 1MiB
    asm.WriteString("intBuf:\n\tresb 12") // int(32bit) -> 10 digits max + \n and sign -> 12 char string max
}


func getArgs(words []word, expectedArgCount int) (args []arg) {
    if len(words) < 2 || words[1].str != "(" {
        fmt.Fprintln(os.Stderr, "[ERROR] missing \"(\"")
        fmt.Fprintln(os.Stderr, "\t" + words[1].at())
        os.Exit(1)
    }

    for _, w := range words[2:] {
        if w.str == ")" {
            break
        }

        if w.str[0] == '"' && w.str[len(w.str) - 1] == '"' {
            args = append(args, arg{false, str, len(strLits)})
            addStrLit(w)
        } else if i, err := strconv.Atoi(w.str); err == nil {
            args = append(args, arg{false, i32, i})
        } else {
            if v := getVar(w.str); v != nil {
                args = append(args, arg{true, v.vartype, v.regIdx})
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", w.str)
                fmt.Fprintln(os.Stderr, "\t" + w.at())
                os.Exit(1)
            }
        }
    }

    if len(words) - 2 == len(args) {
        fmt.Fprintf(os.Stderr, "[ERROR] missing \")\"\n")
        os.Exit(1)
    }

    if len(args) != expectedArgCount {
        fmt.Fprintf(os.Stderr, "[ERROR] function takes %d argument but got %d\n", expectedArgCount, len(args))
        fmt.Fprintln(os.Stderr, "\t" + words[0].at())
        os.Exit(1)
    }

    return args
}

// escape chars (TODO: \n, \t, ...) (done: \\, \")
func split(file string) (words []word) {
    start := 0

    line := 1
    col := 1

    skip := false
    mlSkip := false
    strLit := false
    escape := false

    for i, r := range(file) {
        // comments
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

        // string literales
        } else if strLit {
            if !escape {
                if r == '"' {
                    strLit = false
                } else if r == '\\' {
                    escape = true
                }
            } else {
                escape = false
            }

        } else {
            if r == '"' {       // start string literal
                strLit = true
            }

            if r == '/' {       // start comment
                if file[i+1] == '/' {
                    skip = true
                } else if file[i+1] == '*' {
                    skip = true
                    mlSkip = true
                }

            // split
            } else if unicode.IsSpace(r) || r == '(' || r == ')' {
                if start != i {
                    words = append(words, word{line, col + start - i, file[start:i]})
                }
                start = i + 1

                if r == '(' || r == ')' {
                    words = append(words, word{line, col - 1, string(r)})
                }
            }
        }

        // set word position
        if r == '\n' {
            line++
            col = 0
        }
        col++
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
        switch words[i].str {
        case "var":
            i = declareVar(words, i)
        case ":=":
            i = defineVar(asm, words, i)
        case "println":
            i = write(asm, words, i)

        default:
            fmt.Fprintf(os.Stderr, "[ERROR] keyword \"%s\" is not supported\n", words[i].str)
            fmt.Fprintln(os.Stderr, "\t" + words[i].at())
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
    compile(src)
    // TODO: optimization step

    genExe()
}

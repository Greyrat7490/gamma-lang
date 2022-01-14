package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const SYS_WRITE = 1
const SYS_EXIT = 60

const STDOUT = 1

var strLits []string

func nasm_header(asm *os.File) {
    asm.WriteString("[BITS 64]\n")
    asm.WriteString("section .text\n")
    asm.WriteString("global _start\n")
    asm.WriteString("_start:\n")
}

func nasm_footer(asm *os.File) {
    asm.WriteString("mov rdi, 0\n")
    asm.WriteString(fmt.Sprintf("mov rax, %d\n", SYS_EXIT))
    asm.WriteString("syscall\n")
    asm.WriteString("section .data\n")
    for i, str := range strLits {
        asm.WriteString(fmt.Sprintf("str%d: db \"%s\", 0xa\n", i, str))
    }

    // TODO: .bss section
}

func syscall(asm *os.File, syscallNum uint, args... interface{}) {
    regs := []string{"rdi", "rsi", "rdx", "r10", "r8", "r9"}

    if len(args) > len(regs) {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) syscall only supports %d args\n", len(regs))
        os.Exit(1)
    }

    for i, arg := range args {
        asm.WriteString(fmt.Sprintf("mov %s, %s\n", regs[i], fmt.Sprint(arg)))
    }

    asm.WriteString(fmt.Sprintf("mov rax, %d\n", syscallNum))
    asm.WriteString("syscall\n")
}

func compile(srcFile []byte) {
    asm, err := os.Create("output.asm")
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] could not create \"output.asm\"")
        os.Exit(1)
    }
    defer asm.Close()

    nasm_header(asm)

    words := strings.Fields(string(srcFile))

    for i := 0; i < len(words); i++ {
        switch words[i] {
        case "println":
            if len(words) <= i + 1 {
                fmt.Fprintln(os.Stderr, "[ERROR] you have not provided enough arguments")
                os.Exit(1)
            }

            i++
            syscall(asm, SYS_WRITE, STDOUT, fmt.Sprintf("str%d", len(strLits)), len(words[i]) + 1)
            strLits = append(strLits, words[i])

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

    compile(src)

    genExe()
}


package nasm

import (
    "os"
    "fmt"
    "os/exec"
    "strings"
)

var rodata string
var data string
var bss string

func AddRodata(s string) {
    rodata += s + "\n"
}
func AddData(s string) {
    data += s + "\n"
}
func AddBss(s string) {
    bss += s + "\n"
}

func writeRodata(file *os.File) {
    file.WriteString("\nsection .rodata\n")
    file.WriteString("_true: db \"true\"\n")
    file.WriteString("_false: db \"false\"\n")
    file.WriteString(rodata)
}

func writeBss(file *os.File) {
    file.WriteString("\nsection .bss\n")
    file.WriteString("\tresb 1024 * 1024\n_stack_top:\n") // 1MiB
    file.WriteString("_intBuf: resb 21\n") // max 64bit -> 20 digits max + sign -> 21 char string max
    file.WriteString(bss)
}

func writeData(file *os.File) {
    file.WriteString("\nsection .data\n")
    file.WriteString(data)
}

func Header(file *os.File) {
    file.WriteString("[BITS 64]\n")
    file.WriteString("section .text\n")
    file.WriteString("global _start\n")
}

func Footer(file *os.File) {
    file.WriteString("\n_start:\n")
    file.WriteString("mov rsp, _stack_top\n\n")

    file.WriteString("call main\n")

    file.WriteString("\nmov rdi, 0\n")
    file.WriteString("call exit\n")

    writeRodata(file)
    writeData(file)
    writeBss(file)
}

func GenExe() {
    var stderr strings.Builder

    fmt.Println("[INFO] generating object files...")

    cmd := exec.Command("nasm", "-f", "elf64", "-o", "output.o", "output.asm")
    cmd.Stderr = &stderr
    err := cmd.Run()
    if err != nil {
        s := stderr.String()
        if s == "" {
            fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] %v\n", s)
        }
        os.Exit(1)
    }

    fmt.Println("[INFO] linking object files...")

    cmd = exec.Command("ld", "-o", "output", "output.o")
    cmd.Stderr = &stderr
    err = cmd.Run()
    if err != nil {
        s := stderr.String()
        if s == "" {
            fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] %v\n", s)
        }
        os.Exit(1)
    }

    fmt.Println("[INFO] generated executable")
}

package nasm

import (
    "os"
    "os/exec"
    "fmt"
    "bufio"
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

func writeRodata(file *bufio.Writer) {
    file.WriteString("\nsection .rodata\n")
    file.WriteString("_true: db \"true\"\n")
    file.WriteString("_false: db \"false\"\n")
    file.WriteString(rodata)
}

func writeBss(file *bufio.Writer) {
    file.WriteString("\nsection .bss\n")
    file.WriteString("align 16\n")
    file.WriteString("\tresb 1024 * 1024\n_stack_top:\n") // 1MiB
    file.WriteString(bss)
}

func writeData(file *bufio.Writer) {
    file.WriteString("\nsection .data\n")
    file.WriteString(data)
}

func Header(file *bufio.Writer) {
    file.WriteString("[BITS 64]\n")
    file.WriteString("section .text\n")
    file.WriteString("global _start\n")
}

func Footer(file *bufio.Writer, noMainArg bool) {
    file.WriteString("\n_start:\n")

    if !noMainArg {
        file.WriteString(
            "lea rdi, [rsp+8]\n" +
            "mov rsi, [rsp]\n" +
            "mov rsp, _stack_top\n" +
            "sub rsp, 24\n" +
            "mov [rsp], rdi\n" +
            "mov [rsp+8], rsi\n" +
            "mov [rsp+16], rsi\n")
    } else {
        file.WriteString("mov rsp, _stack_top\n")
    }

    file.WriteString(
        "call main\n" +
        "mov rdi, 0\n" +
        "call exit\n")

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

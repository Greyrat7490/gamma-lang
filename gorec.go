package main

import (
	"fmt"
	"os"
	"io/ioutil"
)

func nasm_header(asm *os.File) {
    asm.WriteString("[BITS 64]\n")
    asm.WriteString("section .text\n")
    asm.WriteString("global _start\n")
    asm.WriteString("_start:\n")
}

func nasm_footer(asm *os.File) {
    asm.WriteString("mov rdi, 0\n")
    asm.WriteString("mov rax, 60\n")
    asm.WriteString("syscall\n")

    // TODO: .bss and .data section
}

func compile(srcFile []byte) {
    asm, err := os.Create("output.asm")
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] could not create \"output.asm\"")
        os.Exit(1)
    }
    
    nasm_header(asm)
 
    // TODO: process src file

    nasm_footer(asm)
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
}


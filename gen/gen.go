package gen

import (
    "os"
    "fmt"
    "gamma/ast"
    "gamma/std"
    "gamma/types/str"
    "gamma/types/array"
    "gamma/gen/asm/x86_64/nasm"
)

func GenAsm(Ast ast.Ast) {
    fmt.Println("[INFO] generating asm x86_64 file...")

    asm, err := os.Create("output.asm")
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] could not create \"output.asm\"")
        os.Exit(1)
    }
    defer asm.Close()

    nasm.Header(asm)

    std.Define(asm)

    for _,d := range Ast.Decls {
        GenDecl(asm, d)
    }

    str.Gen()
    array.Gen()

    nasm.Footer(asm)
}

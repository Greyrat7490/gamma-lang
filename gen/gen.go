package gen

import (
    "os"
    "fmt"
    "bufio"
    "gamma/ast"
    "gamma/buildin"
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
    writer := bufio.NewWriter(asm)

    nasm.Header(writer)

    buildin.Define(writer)

    for _,d := range Ast.Decls {
        GenDecl(writer, d)
    }

    str.Gen()
    array.Gen()

    nasm.Footer(writer, Ast.NoMainArg)

    writer.Flush()
    asm.Close()
}

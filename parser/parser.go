package prs

import (
    "os"
    "fmt"
    "gamma/ast"
    "gamma/token"
    "gamma/import"
    "gamma/buildin"
)

var isMainDefined bool = false
var noMainArg bool = true

func Parse(path string) (ast ast.Ast) {
    fmt.Println("[INFO] parsing...")

    parseBuildin(&ast)
    parseMain(path, &ast)

    return
}

func parseBuildin(ast *ast.Ast) {
    buildin.Declare()

    tokens := imprt.ImportBuildin()
    for tokens.Peek().Type != token.EOF {
        ast.Decls = append(ast.Decls, prsDecl(&tokens))
    }
}

func parseMain(path string, ast *ast.Ast) {
    tokens := imprt.ImportMain(path)

    for tokens.Peek().Type != token.EOF {
        tokens.SetLastImport()
        ast.Decls = append(ast.Decls, prsDecl(&tokens))
    }

    if !isMainDefined {
        fmt.Fprintln(os.Stderr, "[ERROR] no \"main\" function was defined")
        os.Exit(1)
    }

    ast.NoMainArg = noMainArg
}

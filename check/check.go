package check

import (
    "fmt"
    "gamma/ast"
)

func TypeCheck(Ast ast.Ast) {
    fmt.Println("[INFO] typechecking...")

    for _,d := range Ast.Decls {
        typeCheckDecl(d)
    }
}

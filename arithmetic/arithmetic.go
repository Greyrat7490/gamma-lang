package arith

import (
    "fmt"
    "gorec/vars"
    "gorec/token"
    "os"
)

func UnaryOp(operator token.Token, operand token.Token) {
    /* TODO negating vars */
}

func BinaryOp(asm *os.File, opType token.TokenType, src token.Token, dest token.Token) {
    var s string
    var d string

    // TODO get regs method
    if src.Type == token.Name {
        v := vars.GetVar(src.Str)
        if v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", src.Str)
            fmt.Fprintln(os.Stderr, "\t" + src.At())
            os.Exit(1)
        }

        s = vars.Registers[v.Regs[0]].Name
    } else {
        s = src.Str
    }

    if dest.Type == token.Name {
        v := vars.GetVar(dest.Str)
        if v == nil {
            fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" is not declared\n", dest.Str)
            fmt.Fprintln(os.Stderr, "\t" + dest.At())
            os.Exit(1)
        }

        d = vars.Registers[v.Regs[0]].Name
    } else {
        d = dest.Str
    }

    switch opType {
    case token.Plus:
        add(asm, s, d);
    case token.Minus:
        sub(asm, s, d);
    case token.Mul:
        mul(asm, s, d);
    case token.Div:
        div(asm, s, d);
    default:
        fmt.Fprintf(os.Stderr, "Error: only +,-,*,/ are supported binary operators (got %s)\n", opType.Readable())
        os.Exit(1)
    }
}

func add(asm *os.File, src string, dest string) {
    vars.WriteVar(asm, fmt.Sprintf("add %s, %s\n", dest, src))
}

func sub(asm *os.File, src string, dest string) {
    vars.WriteVar(asm, fmt.Sprintf("sub %s, %s\n", dest, src))
}

func mul(asm *os.File, src string, dest string) {
    if src == "rax" && dest == "rbx" {
        vars.WriteVar(asm, "push rax\n")
        vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", dest))
        vars.WriteVar(asm, "pop rbx\n")
        vars.WriteVar(asm, "push rbx\n")

        vars.WriteVar(asm, "imul rbx\n")

        vars.WriteVar(asm, fmt.Sprintf("mov %s, rax\n", dest))
        vars.WriteVar(asm, "pop rax\n")
    } else {
        if dest != "rbx" {
            vars.WriteVar(asm, "push rbx\n")
        }
        if dest != "rax" {
            vars.WriteVar(asm, "push rax\n")
            vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", dest))
        }

        vars.WriteVar(asm, fmt.Sprintf("mov rbx, %s\n", src))
        vars.WriteVar(asm, "imul rbx\n")

        if dest != "rax" {
            vars.WriteVar(asm, fmt.Sprintf("mov %s, rax\n", dest))
            vars.WriteVar(asm, "pop rax\n")
        }
        if dest != "rbx" {
            vars.WriteVar(asm, "pop rbx\n")
        }
    }
}

func div(asm *os.File, src string, dest string) {
    if src == "rax" && dest == "rbx" {
        vars.WriteVar(asm, "push rdx\n")
        vars.WriteVar(asm, "push rax\n")
        vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", dest))
        vars.WriteVar(asm, "pop rbx\n")
        vars.WriteVar(asm, "push rbx\n")

        // TODO: check if dest is signed or unsigned (use either idiv or div)
        // for now only signed integers are supported
        vars.WriteVar(asm, fmt.Sprintf("mov rbx, %s\n", src))
        vars.WriteVar(asm, "cqo\n") // sign extend rax into rdx (div with 64bit regs -> 128bit div)
        vars.WriteVar(asm, "idiv rbx\n")

        vars.WriteVar(asm, fmt.Sprintf("mov %s, rax\n", dest))
        vars.WriteVar(asm, "pop rax\n")
        vars.WriteVar(asm, "pop rdx\n")
    } else {
        if dest != "rdx" {
            vars.WriteVar(asm, "push rdx\n")
        }
        if dest != "rbx" {
            vars.WriteVar(asm, "push rbx\n")
        }
        if dest != "rax" {
            vars.WriteVar(asm, "push rax\n")
            vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", dest))
        }

        // TODO: check if dest is signed or unsigned (use either idiv or div)
        // for now only signed integers are supported
        vars.WriteVar(asm, fmt.Sprintf("mov rbx, %s\n", src))
        vars.WriteVar(asm, "cqo\n") // sign extend rax into rdx (div with 64bit regs -> 128bit div)
        vars.WriteVar(asm, "idiv rbx\n")

        if dest != "rax" {
            vars.WriteVar(asm, fmt.Sprintf("mov %s, rax\n", dest))
            vars.WriteVar(asm, "pop rax\n")
        }
        if dest != "rbx" {
            vars.WriteVar(asm, "pop rbx\n")
        }
        if dest != "rdx" {
            vars.WriteVar(asm, "pop rdx\n")
        }
    }
}

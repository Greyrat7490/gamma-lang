package arith

import (
    "os"
    "fmt"
    "gorec/vars"
    "gorec/token"
)

func BinaryOp(asm *os.File, opType token.TokenType, src string) {
    switch opType {
    case token.Eql:
        eql(asm, "rax", src)
    case token.Neq:
        neq(asm, "rax", src)
    case token.Lss:
        lss(asm, "rax", src)
    case token.Grt:
        grt(asm, "rax", src)
    case token.Leq:
        leq(asm, "rax", src)
    case token.Geq:
        geq(asm, "rax", src)

    case token.Plus:
        add(asm, src, "rax")
    case token.Minus:
        sub(asm, src, "rax")
    case token.Mul:
        mul(asm, src, "rax")
    case token.Div:
        div(asm, src, "rax")
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unknown binary operator %s\n", opType.Readable())
        os.Exit(1)
    }
}

func eql(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "mov rax, 0\n") // only tmp (xor does not work for some reason????)
    vars.Write(asm, "sete al\n")
}
func neq(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "mov rax, 0\n")
    vars.Write(asm, "setne al\n")
}
func lss(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "mov rax, 0\n")
    vars.Write(asm, "setl al\n")
}
func grt(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "mov rax, 0\n")
    vars.Write(asm, "setg al\n")
}
func leq(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "mov rax, 0\n")
    vars.Write(asm, "setle al\n")
}
func geq(asm *os.File, lhs string, rhs string) {
    vars.Write(asm, fmt.Sprintf("cmp %s, %s\n", lhs, rhs))
    vars.Write(asm, "mov rax, 0\n")
    vars.Write(asm, "setge al\n")
}

func add(asm *os.File, src string, dest string) {
    vars.Write(asm, fmt.Sprintf("add %s, %s\n", dest, src))
}

func sub(asm *os.File, src string, dest string) {
    vars.Write(asm, fmt.Sprintf("sub %s, %s\n", dest, src))
}

func mul(asm *os.File, src string, dest string) {
    if src == "rax" && dest == "rbx" {
        vars.Write(asm, "push rax\n")
        vars.Write(asm, fmt.Sprintf("mov rax, %s\n", dest))
        vars.Write(asm, "pop rbx\n")
        vars.Write(asm, "push rbx\n")
        vars.Write(asm, "push rdx\n") // products higher part gets stored in rdx

        vars.Write(asm, "imul rbx\n")

        vars.Write(asm, "pop rdx\n")
        vars.Write(asm, fmt.Sprintf("mov %s, rax\n", dest))
        vars.Write(asm, "pop rax\n")
    } else {
        if dest != "rbx" {
            vars.Write(asm, "push rbx\n")
        }
        if dest != "rax" {
            vars.Write(asm, "push rax\n")
            vars.Write(asm, fmt.Sprintf("mov rax, %s\n", dest))
        }

        vars.Write(asm, "push rdx\n")

        vars.Write(asm, fmt.Sprintf("mov rbx, %s\n", src))
        vars.Write(asm, "imul rbx\n")

        vars.Write(asm, "pop rdx\n")

        if dest != "rax" {
            vars.Write(asm, fmt.Sprintf("mov %s, rax\n", dest))
            vars.Write(asm, "pop rax\n")
        }
        if dest != "rbx" {
            vars.Write(asm, "pop rbx\n")
        }
    }
}

func div(asm *os.File, src string, dest string) {
    if src == "rax" && dest == "rbx" {
        vars.Write(asm, "push rdx\n")
        vars.Write(asm, "push rax\n")
        vars.Write(asm, fmt.Sprintf("mov rax, %s\n", dest))
        vars.Write(asm, "pop rbx\n")
        vars.Write(asm, "push rbx\n")

        // TODO: check if dest is signed or unsigned (use either idiv or div)
        // for now only signed integers are supported
        vars.Write(asm, fmt.Sprintf("mov rbx, %s\n", src))
        vars.Write(asm, "cqo\n") // sign extend rax into rdx (div with 64bit regs -> 128bit div)
        vars.Write(asm, "idiv rbx\n")

        vars.Write(asm, fmt.Sprintf("mov %s, rax\n", dest))
        vars.Write(asm, "pop rax\n")
        vars.Write(asm, "pop rdx\n")
    } else {
        if dest != "rdx" {
            vars.Write(asm, "push rdx\n")
        }
        if dest != "rbx" {
            vars.Write(asm, "push rbx\n")
        }
        if dest != "rax" {
            vars.Write(asm, "push rax\n")
            vars.Write(asm, fmt.Sprintf("mov rax, %s\n", dest))
        }

        // TODO: check if dest is signed or unsigned (use either idiv or div)
        // for now only signed integers are supported
        vars.Write(asm, fmt.Sprintf("mov rbx, %s\n", src))
        vars.Write(asm, "cqo\n") // sign extend rax into rdx (div with 64bit regs -> 128bit div)
        vars.Write(asm, "idiv rbx\n")

        if dest != "rax" {
            vars.Write(asm, fmt.Sprintf("mov %s, rax\n", dest))
            vars.Write(asm, "pop rax\n")
        }
        if dest != "rbx" {
            vars.Write(asm, "pop rbx\n")
        }
        if dest != "rdx" {
            vars.Write(asm, "pop rdx\n")
        }
    }
}

package arithmetic

import (
    "fmt"
    "gorec/parser"
    "gorec/vars"
    "os"
)

func Add(asm *os.File, op *prs.Op) {
    if op.Type != prs.OP_ADD {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_ADD\n")
        os.Exit(1)
    }

    if len(op.Operants) != 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_ADD) should have 2 operants\n")
        os.Exit(1)
    }

    if v := vars.GetVar(op.Operants[0]); v != nil {
        if len(v.Regs) != 1 {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable should have 1 register\n")
            os.Exit(1)
        }

        vars.WriteVar(asm, fmt.Sprintf("add %s, %s\n", vars.Registers[v.Regs[0]].Name, op.Operants[1]))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }
}

func Sub(asm *os.File, op *prs.Op) {
    if op.Type != prs.OP_SUB {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_SUB\n")
        os.Exit(1)
    }

    if len(op.Operants) != 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_SUB) should have 2 operants\n")
        os.Exit(1)
    }

    if v := vars.GetVar(op.Operants[0]); v != nil {
        if len(v.Regs) != 1 {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable should have 1 register\n")
            os.Exit(1)
        }

        vars.WriteVar(asm, fmt.Sprintf("sub %s, %s\n", vars.Registers[v.Regs[0]].Name, op.Operants[1]))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }
}

func Mul(asm *os.File, op *prs.Op) {
    if op.Type != prs.OP_MUL {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_MUL\n")
        os.Exit(1)
    }

    if len(op.Operants) != 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_MUL) should have 2 operants\n")
        os.Exit(1)
    }

    if v := vars.GetVar(op.Operants[0]); v != nil {
        if len(v.Regs) != 1 {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable should have 1 register\n")
            os.Exit(1)
        }

        dest := vars.Registers[v.Regs[0]].Name

        if dest != "rbx" {
            vars.WriteVar(asm, "push rbx\n")
        }
        if dest != "rax" {
            vars.WriteVar(asm, "push rax\n")
            vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", dest))
        }

        vars.WriteVar(asm, fmt.Sprintf("mov rbx, %s\n", op.Operants[1]))
        vars.WriteVar(asm, "mul rbx\n")

        if dest != "rax" {
            vars.WriteVar(asm, fmt.Sprintf("mov %s, rax\n", dest))
            vars.WriteVar(asm, "pop rax\n")
        }
        if dest != "rbx" {
            vars.WriteVar(asm, "pop rbx\n")
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }
}

func Div(asm *os.File, op *prs.Op) {
    if op.Type != prs.OP_DIV {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) OpType should be OP_DIV\n")
        os.Exit(1)
    }

    if len(op.Operants) != 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) Op(OP_DIV) should have 2 operants\n")
        os.Exit(1)
    }

    if v := vars.GetVar(op.Operants[0]); v != nil {
        if len(v.Regs) != 1 {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable should have 1 register\n")
            os.Exit(1)
        }

        dest := vars.Registers[v.Regs[0]].Name

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

        vars.WriteVar(asm, fmt.Sprintf("mov rbx, %s\n", op.Operants[1]))
        vars.WriteVar(asm, "xor rdx, rdx\n")
        vars.WriteVar(asm, "div rbx\n")

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
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }
}

package arithmetic

import (
	"fmt"
	"gorec/parser"
	"gorec/vars"
	"os"
)

func Add(op *prs.Op) {
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

        vars.AddToGlobalScope(fmt.Sprintf("add %s, %s\n", vars.Registers[v.Regs[0]].Name, op.Operants[1]))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }
}

func Sub(op *prs.Op) {
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

        vars.AddToGlobalScope(fmt.Sprintf("sub %s, %s\n", vars.Registers[v.Regs[0]].Name, op.Operants[1]))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }
}

func Mul(op *prs.Op) {
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

        vars.AddToGlobalScope("push rax\npush rbx\n")
        vars.AddToGlobalScope(fmt.Sprintf("mov rax, %s\n", vars.Registers[v.Regs[0]].Name))
        vars.AddToGlobalScope(fmt.Sprintf("mov rbx, %s\n", op.Operants[1]))
        vars.AddToGlobalScope("mul rbx\n")
        vars.AddToGlobalScope(fmt.Sprintf("mov %s, rax\n", vars.Registers[v.Regs[0]].Name))
        vars.AddToGlobalScope("pop rbx\npop rax\n")
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }
}

func Div(op *prs.Op) {
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
            vars.AddToGlobalScope("push rdx\n")
        }

        vars.AddToGlobalScope("push rax\npush rbx\n")
        vars.AddToGlobalScope(fmt.Sprintf("mov rax, %s\n", vars.Registers[v.Regs[0]].Name))
        vars.AddToGlobalScope(fmt.Sprintf("mov rbx, %s\n", op.Operants[1]))
        vars.AddToGlobalScope("xor rdx, rdx\n")
        vars.AddToGlobalScope("div rbx\n")
        vars.AddToGlobalScope(fmt.Sprintf("mov %s, rax\n", vars.Registers[v.Regs[0]].Name))
        vars.AddToGlobalScope("pop rbx\npop rax\n")

        if dest != "rdx" {
            vars.AddToGlobalScope("pop rdx\n")
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", op.Operants[0])
        os.Exit(1)
    }
}

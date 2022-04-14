package arithmetic

import (
    "fmt"
    "gorec/types"
    "gorec/vars"
    "gorec/token"
    "os"
)

// TODO: to one function

func Add(asm *os.File, varname token.Token, value token.Token) {
    if v := vars.GetVar(varname.Str); v != nil {
        if len(v.Regs) != 1 {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable should have 1 register\n")
            os.Exit(1)
        }

        dest := vars.Registers[v.Regs[0]].Name
        var val string

        if otherVar := vars.GetVar(value.Str); otherVar != nil {
            if otherVar.Vartype != types.I32 {
                fmt.Fprintf(os.Stderr, "[ERROR] you cannot add I32 and %s (type of %s)\n", otherVar.Vartype.Readable(), otherVar.Name)
                fmt.Fprintln(os.Stderr, "\t" + value.At())
                os.Exit(1)
            }

            val = vars.Registers[otherVar.Regs[0]].Name
        } else {
            val = value.Str
        }

        vars.WriteVar(asm, fmt.Sprintf("add %s, %s\n", dest, val))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", varname.Str)
        os.Exit(1)
    }
}

func Sub(asm *os.File, varname token.Token, value token.Token) {
    if v := vars.GetVar(varname.Str); v != nil {
        if len(v.Regs) != 1 {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable should have 1 register\n")
            os.Exit(1)
        }

        dest := vars.Registers[v.Regs[0]].Name
        var val string

        if otherVar := vars.GetVar(value.Str); otherVar != nil {
            if otherVar.Vartype != types.I32 {
                fmt.Fprintf(os.Stderr, "[ERROR] you cannot sub I32 and %s (type of %s)\n", otherVar.Vartype.Readable(), otherVar.Name)
                fmt.Fprintln(os.Stderr, "\t" + value.At())
                os.Exit(1)
            }

            val = vars.Registers[otherVar.Regs[0]].Name
        } else {
            val = value.Str
        }

        vars.WriteVar(asm, fmt.Sprintf("sub %s, %s\n", dest, val))
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", varname.Str)
        os.Exit(1)
    }
}

func Mul(asm *os.File, varname token.Token, value token.Token) {
    if v := vars.GetVar(varname.Str); v != nil {
        if len(v.Regs) != 1 {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable should have 1 register\n")
            os.Exit(1)
        }

        dest := vars.Registers[v.Regs[0]].Name
        var val string

        if otherVar := vars.GetVar(value.Str); otherVar != nil {
            if otherVar.Vartype != types.I32 {
                fmt.Fprintf(os.Stderr, "[ERROR] you cannot mul I32 and %s (type of %s)\n", otherVar.Vartype.Readable(), otherVar.Name)
                fmt.Fprintln(os.Stderr, "\t" + value.At())
                os.Exit(1)
            }

            val = vars.Registers[otherVar.Regs[0]].Name
        } else {
            val = value.Str
        }

        if dest != "rbx" {
            vars.WriteVar(asm, "push rbx\n")
        }
        if dest != "rax" {
            vars.WriteVar(asm, "push rax\n")
            vars.WriteVar(asm, fmt.Sprintf("mov rax, %s\n", dest))
        }

        vars.WriteVar(asm, fmt.Sprintf("mov rbx, %s\n", val))
        vars.WriteVar(asm, "imul rbx\n")

        if dest != "rax" {
            vars.WriteVar(asm, fmt.Sprintf("mov %s, rax\n", dest))
            vars.WriteVar(asm, "pop rax\n")
        }
        if dest != "rbx" {
            vars.WriteVar(asm, "pop rbx\n")
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", varname.Str)
        os.Exit(1)
    }
}

func Div(asm *os.File, varname token.Token, value token.Token) {
    if v := vars.GetVar(varname.Str); v != nil {
        if len(v.Regs) != 1 {
            fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable should have 1 register\n")
            os.Exit(1)
        }

        dest := vars.Registers[v.Regs[0]].Name
        var val string

        if otherVar := vars.GetVar(value.Str); otherVar != nil {
            if otherVar.Vartype != types.I32 {
                fmt.Fprintf(os.Stderr, "[ERROR] you cannot div I32 and %s (type of %s)\n", otherVar.Vartype.Readable(), otherVar.Name)
                fmt.Fprintln(os.Stderr, "\t" + value.At())
                os.Exit(1)
            }

            val = vars.Registers[otherVar.Regs[0]].Name
        } else {
            val = value.Str
        }

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
        vars.WriteVar(asm, fmt.Sprintf("mov rbx, %s\n", val))
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
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) variable \"%s\" is not declared\n", varname.Str)
        os.Exit(1)
    }
}
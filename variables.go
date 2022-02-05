package main

import (
    "fmt"
    "os"
    "strconv"
)

type reg struct {
    name string
    isAddr bool
    value int      // either an actual value or an address(index)
}

// TODO: register allocator for variables
var registers []reg = []reg {
    {name: "rax"},
    {name: "rbx"},
    {name: "rcx"},
    {name: "rdx"},
    {name: "r8"},
    {name: "r9"},
    {name: "r10"},
    {name: "r11"},
}

const maxRegs int = 5
var availReg int = 0

var vars []variable
var globalDefs []string

type variable struct {
    name string
    regs []int
    vartype gType
}

func getVar(varname string) *variable {
    for _, v := range vars {
        if v.name == varname {
            return &v
        }
    }

    return nil
}

func isLit(w string) bool {
    if w[0] == '"' && w[len(w) - 1] == '"' {
        return true
    }

    if _, err := strconv.Atoi(w); err == nil {
        return true
    }

    return false
}

func declareVar(words []word, i int) int {
    if len(words) < i + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] neither name nor type provided for the variable declaration")
        fmt.Fprintln(os.Stderr, "\t" + words[i].at())
        os.Exit(1)
    }
    if len(words) < i + 2 {
        fmt.Fprintln(os.Stderr, "[ERROR] no name or type provided for the variable")
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
        os.Exit(1)
    }
    if len(vars) >= len(registers) {
        fmt.Fprintf(os.Stderr, "[ERROR] a maximum of only %d variables is allowed yet\n", len(registers))
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
        os.Exit(1)
    }
    // maybe implement shadowing later (TODO)
    if getVar(words[i+1].str) != nil {
        fmt.Fprintf(os.Stderr, "[ERROR] a variable with the name \"%s\" is already declared\n", words[i+1].str)
        fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
        os.Exit(1)
    }

    switch toType(words[i+2].str) {
    case str:
        if availReg + 1 >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(string)", words[i+1].str)
            fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
            os.Exit(1)
        }

        vars = append(vars, variable{words[i+1].str, []int{ availReg, availReg+1 }, str})
        availReg += 2
    case i32:
        if availReg >= maxRegs {
            fmt.Fprintf(os.Stderr, "[ERROR] not enough registers left for var \"%s\"(i32)", words[i+1].str)
            fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
            os.Exit(1)
        }

        vars = append(vars, variable{words[i+1].str, []int{ availReg }, i32})
        availReg++
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", words[i+2].str)
        fmt.Fprintln(os.Stderr, "\t" + words[i+2].at())
        os.Exit(1)
    }

    return i + 2
}

func defineVar(words []word, i int) int {
    if len(words) < i + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + words[i].at())
        os.Exit(1)
    }

    if isLit(words[i+1].str) {
        if v := getVar(words[i-2].str); v != nil {
            switch v.vartype {
            case str:
                if len(v.regs) != 2 {
                    fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) string variable should use 2 registers\n")
                    fmt.Fprintln(os.Stderr, "\t" + words[i-2].at())
                    os.Exit(1)
                }

                globalDefs = append(globalDefs, fmt.Sprintf("mov %s, str%d\n", registers[v.regs[0]].name, len(strLits)))
                addStrLit(words[i+1].str)
                globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %d\n", registers[v.regs[1]].name, strLits[len(strLits)-1].size))

            case i32:
                i, _ := strconv.Atoi(words[i+1].str)
                globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %d\n", registers[v.regs[0]].name, i))

            default:
                fmt.Fprintf(os.Stderr, "[ERROR] (unreachable) the type of \"%s\" is not set correctly\n", v.name)
                fmt.Fprintln(os.Stderr, "\t" + words[i-2].at())
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" not declared\n", words[i-2].str)
            fmt.Fprintln(os.Stderr, "\t" + words[i-2].at())
            os.Exit(1)
        }
    } else {
        // TODO: check if var is defined
        if otherVar := getVar(words[i+1].str); otherVar != nil {
            if v := getVar(words[i-2].str); v != nil {
                for ri, r := range otherVar.regs {
                    globalDefs = append(globalDefs, fmt.Sprintf("mov %s, %s\n", registers[v.regs[ri]].name, registers[r].name))
                }
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" not declared\n", words[i-2].str)
                fmt.Fprintln(os.Stderr, "\t" + words[i-2].at())
                os.Exit(1)
            }
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not declared\n", words[i+1].str)
            fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
            os.Exit(1)
        }
    }

    return i + 1
}

func rmVar(varname string) {
    if len(vars) == 1 && vars[0].name == varname {
        vars = []variable{}
        return
    }

    for i, v := range vars {
        if v.name == varname {
            vars[i] = vars[len(vars)-1]
            vars = vars[:len(vars)-1]
            return
        }
    }
}

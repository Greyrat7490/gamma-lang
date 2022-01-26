package main

import (
    "fmt"
    "os"
    "strconv"
)

var vars []variable

type variable struct {
    name string
    regIdx int
    vartype gType
    strIdx int
}

func getVar(varname string) *variable {
    for _, v := range vars {
        if v.name == varname {
            return &v
        }
    }

    return nil
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
        vars = append(vars, variable{words[i+1].str, len(vars), str, -1})
    case i32:
        vars = append(vars, variable{words[i+1].str, len(vars), i32, -1})
    default:
        fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid type\n", words[i+2].str)
        fmt.Fprintln(os.Stderr, "\t" + words[i+2].at())
        os.Exit(1)
    }

    return i + 2
}

func defineVar(asm *os.File, words []word, i int) int {
    if len(words) < i + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        fmt.Fprintln(os.Stderr, "\t" + words[i].at())
        os.Exit(1)
    }

    if otherVar := getVar(words[i+1].str); otherVar != nil {        // define with variable
        if v := getVar(words[i-2].str); v != nil {
            registers[v.regIdx].isAddr = registers[otherVar.regIdx].isAddr;
            registers[v.regIdx].value = registers[otherVar.regIdx].value;
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", registers[v.regIdx].name, registers[otherVar.regIdx].name))
        } else {
            fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" not declared\n", words[i-2].str)
            fmt.Fprintln(os.Stderr, "\t" + words[i-2].at())
            os.Exit(1)
        }
    } else {                                                        // define with literal
        if v := getVar(words[i-2].str); v != nil {
            switch v.vartype {
            case str:
                registers[v.regIdx].isAddr = true;
                registers[v.regIdx].value = len(strLits);

                strLits = append(strLits, words[i+1].str)
                asm.WriteString(fmt.Sprintf("mov %s, %s\n", registers[v.regIdx].name, fmt.Sprintf("str%d", registers[v.regIdx].value)))

            case i32:
                registers[v.regIdx].isAddr = false;

                if i, err := strconv.Atoi(words[i+1].str); err == nil {
                    registers[v.regIdx].value = i;

                    asm.WriteString(fmt.Sprintf("mov %s, %d\n", registers[v.regIdx].name, i))
                } else {
                    fmt.Fprintf(os.Stderr, "[ERROR] \"%s\" is not a valid integer\n", words[i+1].str)
                fmt.Fprintln(os.Stderr, "\t" + words[i+1].at())
                    os.Exit(1)
                }

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
    }

    return i + 1
}


package main

import (
    "fmt"
    "os"
)

var vars []variable

type variable struct {
    name string
    regIdx int
    vartype vType
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

func declareVar(words []string, i int) int {
    if len(words) < i + 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] neither name nor type provided for the variable declaration")
        os.Exit(1)
    }
    if len(words) < i + 2 {
        fmt.Fprintln(os.Stderr, "[ERROR] no name or type provided for the variable")
        os.Exit(1)
    }
    if len(vars) >= len(registers) {
        fmt.Fprintf(os.Stderr, "[ERROR] a maximum of only %d variables is allowed yet\n", len(registers))
        os.Exit(1)
    }

    if words[i+2] == "str" {
        vars = append(vars, variable{words[i+1], len(vars), String, -1})
    } else if words[i+2] == "int" {
        vars = append(vars, variable{words[i+1], len(vars), Int, -1})
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] only str and int are supported yet\n")
        os.Exit(1)
    }

    return i + 2
}

// define by other variable
func defineVar(asm *os.File, words []string, i int) int {
    if len(words) < i + 1 {
        fmt.Fprintf(os.Stderr, "[ERROR] no value provided to define the variable\n")
        os.Exit(1)
    }
    
    if v := getVar(words[i-2]); v != nil {
        switch v.vartype {
        case String:
            registers[v.regIdx].isAddr = true;
            registers[v.regIdx].strIdx = len(strLits);

            strLits = append(strLits, words[i+1])
 
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", registers[v.regIdx].name, fmt.Sprintf("str%d", registers[v.regIdx].strIdx)))

        case Int:
            registers[v.regIdx].isAddr = false;
            registers[v.regIdx].value = words[i+1];
            asm.WriteString(fmt.Sprintf("mov %s, %s\n", registers[v.regIdx].name, registers[v.regIdx].value))

        default:
            // TODO: type to human readable
            fmt.Fprintf(os.Stderr, "[ERROR] \"%#v\" is not supported, only str and int are supported yet\n", v.vartype)
            os.Exit(1)
        }
    } else {
        fmt.Fprintf(os.Stderr, "[ERROR] var \"%s\" not declared\n", words[i-2])
        os.Exit(1)
    }
    
    return i + 1
}


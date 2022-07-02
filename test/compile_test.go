package test

import (
    "os"
    "fmt"
    "flag"
    "os/exec"
    "strings"
    "testing"
    "io/ioutil"
    "path/filepath"
)

var rec bool
var keepAsm bool

var failed bool = false

func record(t *testing.T, name string, stdout string, stderr string) {
    output := []byte(stdout + "\n" + stderr)

    err := ioutil.WriteFile(name, output, 0644)
    if err != nil {
        t.Fatalf("[ERROR] could not record results\n\t%v\n", err)
    }

    fmt.Println("[RECORDED]")
}

func check(name string, stdout string, stderr string) {
    result := stdout + "\n" + stderr

    expected, err := ioutil.ReadFile(name)
    if err != nil {
        fmt.Printf("[ERROR] could not compair with recorded results\n\t%v\n", err)
        failed = true
        return
    }

    if result != string(expected) {
        fmt.Println("[FAILED]")
        fmt.Println("--------------------")

        fmt.Fprintln(os.Stderr, "result:")
        fmt.Fprint(os.Stderr, result)

        fmt.Println("-----")

        fmt.Fprintln(os.Stderr, "expected:")
        fmt.Fprint(os.Stderr, string(expected))
        fmt.Println("--------------------")

        failed = true
    } else {
        fmt.Println("[PASSED]")
    }
}

// removes executable, object and assembly files
// if 'srcname' != "" -> the assembly file will be renamed instead
func clearBuilds(t *testing.T, srcname string) {
    err := os.Remove("output")
    if err != nil && !os.IsNotExist(err) {
        t.Fatalf("[ERROR] could not remove output\n%v", err)
    }

    err = os.Remove("output.o")
    if err != nil && !os.IsNotExist(err) {
        t.Fatalf("[ERROR] could not remove output.o\n%v", err)
    }

    if len(srcname) == 0 {
        err = os.Remove("output.asm")
        if err != nil && !os.IsNotExist(err)  {
            t.Fatalf("[ERROR] could not remove output.asm\n%v", err)
        }
    } else {
        err = os.Rename("output.asm", srcname + ".asm")
        if err != nil && !os.IsNotExist(err) {
            t.Fatalf("[ERROR] could not rename output.asm\n%v", err)
        }
    }
}

func init() {
    flag.BoolVar(&rec, "rec", false, "record the stdout and stderr results")
    flag.BoolVar(&keepAsm, "asm", false, "keep the assembly files generated")
}

func TestCompile(t *testing.T) {
    flag.Parse()

    path, err := os.Getwd()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    files, err := ioutil.ReadDir(path)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    for _, f := range files {
        if filepath.Ext(f.Name()) == ".gore" {
            fmt.Print(f.Name())
            cmd := exec.Command("go", "run", "gorec", f.Name())

            var stdout, stderr strings.Builder
            cmd.Stdout = &stdout
            cmd.Stderr = &stderr

            cmd.Run()

            stdoutStr := stdout.String()
            stderrStr := stderr.String()

            if rec {
                record(t, f.Name() + ".rec", stdoutStr, stderrStr)
            } else {
                check(f.Name() + ".rec", stdoutStr, stderrStr)
            }

            if keepAsm {
                clearBuilds(t, f.Name())
            } else {
                clearBuilds(t, "")
            }
        }
    }

    if failed {
        os.Exit(1)
    }
}

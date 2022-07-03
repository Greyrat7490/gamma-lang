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

func record(t *testing.T, path string, name string, stdout string, stderr string) {
    os.MkdirAll(path, 0644)

    output := []byte(stdout + "\n" + stderr)
    file := path + "/" + name + ".rec"

    err := ioutil.WriteFile(file, output, 0644)
    if err != nil {
        t.Fatalf("[ERROR] could not record results\n\t%v\n", err)
    }

    fmt.Println("[RECORDED]")
}

func check(path string, name string, stdout string, stderr string) {
    result := stdout + "\n" + stderr
    file := path + "/" + name + ".rec"

    expected, err := ioutil.ReadFile(file)
    if err != nil {
        fmt.Printf("[ERROR] could not compare with recorded results\n\t%v\n", err)
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
func clearBuilds(t *testing.T, path string, srcname string) {
    err := os.Remove(path + "output")
    if err != nil && !os.IsNotExist(err) {
        t.Fatalf("[ERROR] could not remove output\n%v", err)
    }

    err = os.Remove(path + "output.o")
    if err != nil && !os.IsNotExist(err) {
        t.Fatalf("[ERROR] could not remove output.o\n%v", err)
    }

    if len(srcname) == 0 {
        err = os.Remove(path + "output.asm")
        if err != nil && !os.IsNotExist(err)  {
            t.Fatalf("[ERROR] could not remove output.asm\n%v", err)
        }
    } else {
        err = os.Rename(path + "output.asm", srcname + ".asm")
        if err != nil && !os.IsNotExist(err) {
            t.Fatalf("[ERROR] could not rename output.asm\n%v", err)
        }
    }
}

func init() {
    flag.BoolVar(&rec, "rec", false, "record the stdout and stderr results")
    flag.BoolVar(&keepAsm, "asm", false, "keep the assembly files generated")
}

func test(t *testing.T, flagStr string, recDir string, fileDir string) {
    flag.Parse()

    path, err := os.Getwd()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    path += "/" + fileDir + "/"

    files, err := ioutil.ReadDir(path)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    for _, f := range files {
        if filepath.Ext(f.Name()) == ".gore" {
            fmt.Print(f.Name())
            cmd := exec.Command("go", "run", "gorec", flagStr, f.Name())

            var stdout, stderr strings.Builder
            cmd.Stdout = &stdout
            cmd.Stderr = &stderr

            cmd.Run()

            stdoutStr := stdout.String()
            stderrStr := stderr.String()

            if rec {
                record(t, recDir, f.Name(), stdoutStr, stderrStr)
            } else {
                check(recDir, f.Name(), stdoutStr, stderrStr)
            }

            if keepAsm {
                clearBuilds(t, path, f.Name())
            } else {
                clearBuilds(t, path, "")
            }
        }
    }
}

func TestError(t *testing.T) {
    test(t, "", "recs/err", "errTests")

    if failed { os.Exit(1) }
}

func TestAst(t *testing.T) {
    test(t, "-ast", "recs/ast", "")
}

func TestRun(t *testing.T) {
    test(t, "-r", "recs/run", "")
}

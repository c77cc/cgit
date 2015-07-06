package main

import(
    "fmt"
    "flag"
    "strconv"
    "os"
    "bytes"
    "strings"
	"os/exec"
    fcolor "github.com/fatih/color"
)

var GIT_BIN string = "/usr/bin/git"
var path, _ = os.Getwd()
var committedStr string = "Changes to be committed:"
var notStagedCommitStr string = "Changes not staged for commit:"
var untrackedStr string = "Untracked files:"

func main() {
    if len(os.Args) == 1 {
        usage()
        return
    }
    if len(os.Args) > 1 {
        c := os.Args[1]
        if c == "-h" || c == "--help" || c == "help" {
            usage()
            return
        }
    }

    flag.Parse()
    cmd := flag.Arg(0)
    execCommand(cmd)
}

func usage() {
    help := fmt.Sprintf("Usage: %s [command] [OPTIONS]\n\n", os.Args[0])
    help += fmt.Sprintf("command:\n")
    help += fmt.Sprintf("\tstatus, st\tshow the working tree status\n")
    help += fmt.Sprintf("\tcheckout, co\tcheckout a branch or paths to the working tree\n")
    help += fmt.Sprintf("\treset, re\treset current HEAD to the specified state\n")
    help += fmt.Sprintf("\tadd\t\tadd file contents to the index\n")
    help += fmt.Sprintf("OPTIONS:\n")
    help += fmt.Sprintf("\t0,1\t\t0, 1 files will be apply your commmand\n")
    help += fmt.Sprintf("\t0-2\t\t0, 1, 2 files will be apply your commmand\n")
    help += fmt.Sprintf("\t0-2,3\t\t0, 1, 2, 3 files will be apply your commmand\n")
    help += fmt.Sprintf("\tfilename\tfile will be apply your commmand\n")
    fmt.Println(help)
}

func execCommand(cmd string) {
    var out []byte
    switch cmd {
    case "status", "st":
        out, _, _, _ = execStatus()
    case "checkout", "co":
        out = execCheckout()
    case "reset", "re":
        out = execReset()
    case "add":
        out = execAdd()
    default:
        out = []byte("Unsupported command: " + cmd)
    }
    fmt.Println(string(out))
}

func execStatus() (status []byte, committed, notStagedCommit, untracked []string) {
    out := doExecCommand("status")

    var (
        desc []string
        descs [][]string
        idx int
        state string
        newDesc bool = true
    )

    ret := strings.Split(string(out), "\n")
    for i, line := range ret {
        switch line {
        case committedStr:
            state = "committed"
        case notStagedCommitStr:
            state = "notStagedCommit"
        case untrackedStr:
            state = "untracked"
        default:
        }

        if strings.Index(line, "\t") == 0 {
            trimLine := strings.TrimSpace(strings.Trim(ret[i], "\t"))
            switch state {
            case "committed":
                committed = append(committed, fmt.Sprintf("\t[%d]\t%s", idx, trimLine))
                idx++
            case "notStagedCommit":
                notStagedCommit = append(notStagedCommit, fmt.Sprintf("\t[%d]\t%s", idx, trimLine))
                idx++
            case "untracked":
                untracked = append(untracked, fmt.Sprintf("\t[%d]\t%s", idx, trimLine))
                idx++
            }
            newDesc = true
            continue
        }

        if newDesc {
            if len(desc) > 0 {
                descs = append(descs, desc)
            }

            desc = []string{}
        }

        desc = append(desc, line)
        newDesc = false
    }
    if len(descs) < 1 && len(desc) > 0 {
        descs = append(descs, desc)
    }

    var str string
    for i, _ := range descs {
        str += makeDesc(descs[i])
        if strings.Contains(strings.Join(descs[i], ""), committedStr) {
            str += makeDesc(committed, "yellow")
        }
        if strings.Contains(strings.Join(descs[i], ""), notStagedCommitStr) {
            str += makeDesc(notStagedCommit, "green")
        }
        if strings.Contains(strings.Join(descs[i], ""), untrackedStr) {
            str += makeDesc(untracked, "cyan")
        }
    }

    status = []byte(str)
    return
}

func execCheckout() (s []byte) {
    args := flag.Args()
    idxArr := parseIndexOpts(flag.Arg(1))
    if len(idxArr) < 1 {
        return doExecCommand(args...)
    }

    var fileArr []string
    _, committed, notStagedCommit, _ := execStatus()
    arr := mergeStrArr(committed, notStagedCommit)
    for _, line := range arr {
        idx, ok := getLineNum(line)
        if !ok {
            continue
        }
        if inIntArray(idx, idxArr) {
            if file := getFilepath(line); len(file) > 0 {
                fileArr = append(fileArr, path + "/" + file)
            }
        }
    }

    if len(fileArr) < 1 {
        return doExecCommand(args...)
    }

    for _, file := range fileArr {
        doExecCommand("checkout", file)
    }
    return
}

func execReset() (s []byte) {
    if flag.Arg(1) != "HEAD" {
        fmt.Println("HEAD miss")
    }

    args := flag.Args()
    idxArr := parseIndexOpts(flag.Arg(2))
    if len(idxArr) < 1 {
        return doExecCommand(args...)
    }
    var fileArr []string
    _, committed, _, _ := execStatus()
    for _, line := range committed {
        idx, ok := getLineNum(line)
        if !ok {
            continue
        }
        if inIntArray(idx, idxArr) {
            if file := getFilepath(line); len(file) > 0 {
                fileArr = append(fileArr, path + "/" + file)
            }
        }
    }
    fmt.Println(fileArr)

    if len(fileArr) < 1 {
        return doExecCommand(args...)
    }

    for _, file := range fileArr {
        doExecCommand("reset", "HEAD", file)
    }
    return
}

func execAdd() (s []byte) {
    args := flag.Args()
    idxArr := parseIndexOpts(flag.Arg(1))
    if len(idxArr) < 1 {
        return doExecCommand(args...)
    }
    var fileArr []string
    _, _, notStagedCommit, untracked := execStatus()
    arr := mergeStrArr(notStagedCommit, untracked)
    for _, line := range arr {
        idx, ok := getLineNum(line)
        if !ok {
            continue
        }
        if inIntArray(idx, idxArr) {
            if file := getFilepath(line); len(file) > 0 {
                fileArr = append(fileArr, path + "/" + file)
            }
        }
    }
    fmt.Println(fileArr)

    if len(fileArr) < 1 {
        return doExecCommand(args...)
    }

    for _, file := range fileArr {
        doExecCommand("add", file)
    }
    return
}

func doExecCommand(args ...string) (o []byte) {
    command := exec.Command(GIT_BIN, args...)
    var (
        out bytes.Buffer
        stderr bytes.Buffer
    )

    command.Stdout = &out
    command.Stderr = &stderr
    err := command.Run()
    if err != nil {
        fmt.Println(string(stderr.Bytes()))
        return
    }
    return out.Bytes()
}

func stringToInt(arg string) int {
    i, _ := strconv.ParseInt(arg, 10, 32)
    return int(i)
}

func isNumberic(arg string) bool {
    _, err := strconv.Atoi(arg)
    return err == nil
}

func inIntArray(find int, arr []int) (ok bool) {
    for i, _ := range arr {
        if find == arr[i] {
            return true
        }
    }
    return
}

func parseIndexOpts(opts string) (idxArr []int) {
    optsArr := strings.Split(opts, ",")
    if len(optsArr) < 1 {
        return
    }
    for i, _ := range optsArr {
        arr := strings.Split(optsArr[i], "-")
        if len(arr) < 1 {
            continue
        }
        if len(arr) == 1 && isNumberic(optsArr[i]) {
            idxArr = append(idxArr, stringToInt(optsArr[i]))
            continue
        }
        if len(arr) == 2 && isNumberic(arr[0]) && isNumberic(arr[1]) {
            begin := stringToInt(arr[0])
            end   := stringToInt(arr[1])
            for k := begin; k <= end; k++ {
                idxArr = append(idxArr, k)
            }
            continue
        }
    }
    return
}

func makeDesc(desc []string, colors ...string) (str string) {
    var color string
    if len(colors) > 0 {
        color = colors[0]
    }

    for i, _ := range desc {
        switch color {
        case "green":
            str += fcolor.GreenString(desc[i]) + "\n"
        case "yellow":
            str += fcolor.YellowString(desc[i]) + "\n"
        case "cyan":
            str += fcolor.CyanString(desc[i]) + "\n"
        default:
            str += desc[i] + "\n"
        }
    }
    return
}

func getFilepath(line string) (file string) {
    strArr := strings.Split(strings.TrimSpace(strings.Trim(line, "\t")), "\t")
    fileArr := strings.Split(strArr[len(strArr)-1], ":")
    file = strings.TrimSpace(fileArr[len(fileArr)-1])
    return
}

func getLineNum(line string) (num int, ok bool) {
    newline := strings.TrimLeft(line, "\t")
    idx := newline[1:2]
    if len(idx) < 1 {
        return
    }
    return stringToInt(idx), true
}

func mergeStrArr(a []string, b []string, args ...[]string) (arr []string) {
    arr = append(arr, a...)
    arr = append(arr, b...)

    if len(args) > 0 {
        for i, _ := range args {
            arr = append(arr, args[i]...)
        }
    }

    return arr
}

package main

import(
    "fmt"
    "flag"
    "strconv"
    //"log"
    "os"
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
    help += fmt.Sprintf("\tall git support commands\n")
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
        out = doExecCommand(flag.Args()...)
    }
    fmt.Println(string(out))
}

func execStatus() (status []byte, committed, notStagedCommit, untracked []string) {
    out := doExecCommand("status")

    var desc []string
    var descs [][]string
    var idx int
    var state string
    var newDesc bool = true

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
    for i, line := range notStagedCommit {
        idx := len(committed) + i
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
    for i, line := range committed {
        idx := i
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
    _, committed, notStagedCommit, untracked := execStatus()
    for i, line := range untracked {
        idx := len(committed) + len(notStagedCommit) + i
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
    out, err := command.Output()
    if err != nil {
        fmt.Println(err.Error())
        return out
    }
    return out
}

func stringToInt(arg string) int {
    i, _ := strconv.ParseInt(arg, 10, 32)
    return int(i)
}

func isNumberic(arg string) bool {
    _, err := strconv.Atoi(arg)
    return err == nil
}

func inIntArray(find int, arr []int) bool {
    for i, _ := range arr {
        if find == arr[i] {
            return true
        }
    }
    return false
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
            //fmt.Println(optsArr[i] + " invalid, skiped")
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
    str := strings.TrimSpace(strings.Trim(line, "\t"))
    strArr := strings.Split(str, "\t")
    file = strArr[len(strArr)-1]
    fileArr := strings.Split(file, ":")
    file = fileArr[len(fileArr)-1]

    file = strings.TrimSpace(file)
    return
}

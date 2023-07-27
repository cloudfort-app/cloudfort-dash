package main

import (
    "bufio"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
    "time"

    "github.com/Jeffail/gabs/v2"

    "github.com/gorilla/websocket"

    "github.com/creack/pty"
)

var config, repos *gabs.Container
var domain string
var home string
var port string
var password []byte
var version = "v0.1.28"

var upgrader = websocket.Upgrader{}

/*func check_referer(req *http.Request) bool {
    return (req.Referer() == "" || req.Referer()[0:len(domain)] != domain)
}*/

func fileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}

func hasValidWebhookSignature(w *http.ResponseWriter, req *http.Request, msg []byte) bool {
    mac := hmac.New(sha256.New, password)
    mac.Write(msg);

    if("sha256=" + hex.EncodeToString(mac.Sum(nil)) == req.Header.Get("X-Hub-Signature-256")) {
        fmt.Println("request has valid signature")
        return true;
    } else {
        fmt.Println("request has invalid signature")
        (*w).Header().Set("Content-Type", "text/plain")
        (*w).Write([]byte("invalid signature"))

        return false;
    }
}

func signatureHandshake(conn *websocket.Conn) bool {
    _, date, _ := conn.ReadMessage()
    _, signature, _ := conn.ReadMessage()

    mac := hmac.New(sha256.New, password)
    mac.Write([]byte(date))

    req_date, _ := strconv.Atoi(string(date))

    if(req_date - int(time.Now().UnixMilli()) > 10*1000) {
        fmt.Println("handshake is from the future")
        conn.WriteMessage(websocket.TextMessage, []byte("invalid signature"))
        conn.Close()

        return false;
    } else if(int(time.Now().UnixMilli()) - req_date > 10*1000) {
        fmt.Println("handshake is older than 10 seconds")
        conn.WriteMessage(websocket.TextMessage, []byte("invalid signature"))
        conn.Close()

        return false;
    } else if(hex.EncodeToString(mac.Sum(nil)) != string(signature)) {
        fmt.Println("handshake has invalid signature")
        conn.WriteMessage(websocket.TextMessage, []byte("invalid signature"))
        conn.Close()

        return false;
    } else {
        //fmt.Println("valid request")
        return true;
    }
}

func hasValidSignature(w *http.ResponseWriter, req *http.Request) bool {
    mac := hmac.New(sha256.New, password)
    mac.Write([]byte(req.PostFormValue("date")))

    req_date, _ := strconv.Atoi(req.PostFormValue("date"))

    if(req_date - int(time.Now().UnixMilli()) > 10*1000) {
        fmt.Println("request is from the future")
        (*w).Header().Set("Content-Type", "text/plain")
        (*w).Write([]byte("invalid signature"))

        return false;
    } else if(int(time.Now().UnixMilli()) - req_date > 10*1000) {
        fmt.Println("request is older than 10 seconds")
        (*w).Header().Set("Content-Type", "text/plain")
        (*w).Write([]byte("invalid signature"))

        return false;
    } else if(hex.EncodeToString(mac.Sum(nil)) != req.Header.Get("signature")) {
        fmt.Println("request has invalid signature")
        (*w).Header().Set("Content-Type", "text/plain")
        (*w).Write([]byte("invalid signature"))

        return false;
    } else {
        //fmt.Println("valid request")
        return true;
    }
}

func sanitize(in_str string) string {
    str := strings.Replace(in_str, "\r\n", "\n", -1)

    m1 := regexp.MustCompile(`\n.*\r`)
    str = m1.ReplaceAllString(str, "\n")

    str = strings.Replace(str, "\\", "\\\\", -1)

    //str = strings.Replace(str, "\n", "<br>", -1)
    str = strings.Replace(str, "\n", "\\n", -1)
    str = strings.Replace(str, "\t", "\\t", -1)

    str = strings.Replace(str, "\b", "\\b", -1)
    //str = strings.Replace(str, "\e", "\\e", -1)
    str = strings.Replace(str, "\f", "\\f", -1)
    str = strings.Replace(str, "\r", "\\r", -1)

    str = strings.Replace(str, "\"", "\\\"", -1)

    return str;
}

func sanitize_for_saving(in_str string) string {
    str := strings.Replace(in_str, "\r\n", "\n", -1)

    return str;
}

func responseStr(result bool, response string) string {
    return `{"successful": ` + strconv.FormatBool(result) + `, "response": "` + response + `"}`  
}

func respond(w *http.ResponseWriter, result bool, response string) {
    (*w).Header().Set("Content-Type", "text/plain")
    fmt.Println(responseStr(result, response))
    fmt.Fprintf(*w, responseStr(result, response))
}

func responseJSONStr(result bool, response string) string {
    return `{"successful": ` + strconv.FormatBool(result) + `, "response": ` + response + `}`  
}

func respondJSON(w *http.ResponseWriter, result bool, response string) {
    (*w).Header().Set("Content-Type", "text/plain")
    fmt.Println(responseJSONStr(result, response))
    fmt.Fprintf(*w, responseJSONStr(result, response))
}

func routeCd(w http.ResponseWriter, req *http.Request) {
    /*if(check_referer(req)) {
        w.Write([]byte("404 page not found"))
        return
    }*/

    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""

    err := os.Chdir(req.PostFormValue("dir"))
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routeDirCreate(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""

    err := os.Mkdir(req.PostFormValue("path"), 0755)
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    } 

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routeDeploy(w http.ResponseWriter, req *http.Request) {
    fmt.Println("received deploy request");

    body_bytes, _ := io.ReadAll(req.Body);

    if(!hasValidWebhookSignature(&w, req, body_bytes)) {
        return
    }
    
    body, _ := gabs.ParseJSON(body_bytes)
    full_name, _ := body.Path("repository.full_name").Data().(string)
    org := strings.Split(full_name, "/")[0]
    name := strings.Split(full_name, "/")[1]
    repo_path, _ := repos.Path("repositories." + org + "." + name + ".path").Data().(string)
    deploy_cmd, _ := repos.Path("repositories." + org + "." + name + ".deploy").Data().(string)
    out, err := exec.Command("/bin/sh", "-c", "cd " + repo_path + "; " + deploy_cmd).CombinedOutput()

    if(err != nil) {
        log.Println(err)
    } else {
        fmt.Println(string(out))
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("ok"))
}

func routeFileCreate(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""

    file, err := os.Create(req.PostFormValue("path"))
    if(err != nil) {
        log.Println(err)
        output = sanitize(err.Error() + "\\n")
    } else {
        defer file.Close()
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routeFileRead(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    b, err := os.ReadFile(req.PostFormValue("path")) 
    if err != nil {
        log.Println(err)
    }

    //str := sanitize(string(b));
    str := string(b);

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"content\": \"" + str + "\"}"))
    w.Write([]byte(str))
    //DO NOT USE fmt (causes problems with % characters)
}

func routeFileWrite(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""

    err := os.WriteFile(req.PostFormValue("path"), []byte(sanitize_for_saving(req.PostFormValue("content"))), 0644);
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routeHome(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(home))
}

func routeRename(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""
    err := os.Rename(req.PostFormValue("path-old"), req.PostFormValue("path-new"))
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routeMv(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""
    paths := strings.Split(req.PostFormValue("paths"), " ")

    for p:=0; p<len(paths); p++ {
        sl := strings.Split(paths[p], "/")
        err := os.Rename(paths[p], req.PostFormValue("path") + "/" + sl[len(sl)-1])
        if(err != nil) {
            log.Println(err)
            output = err.Error() + "\\n"
        }
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func Copy(srcpath, dstpath string) (err error) {
    r, err := os.Open(srcpath)
    if err != nil {
        return err
    }
    defer r.Close() // ignore error: file was opened read-only.

    w, err := os.Create(dstpath)
    if err != nil {
        return err
    }

    defer func() {
        // Report the error from Close, if any,
        // but do so only if there isn't already
        // an outgoing error.
        if c := w.Close(); c != nil && err == nil {
                err = c
        }
    }()

    _, err = io.Copy(w, r)
    return err
}

func routeCp(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""
    paths := strings.Split(req.PostFormValue("paths"), " ")

    for p:=0; p<len(paths); p++ {
        sl := strings.Split(paths[p], "/")
        err := Copy(paths[p], req.PostFormValue("path") + "/" + sl[len(sl)-1])
        if(err != nil) {
            log.Println(err)
            output = err.Error() + "\\n"
        }
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routeRm(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""
    paths := strings.Split(req.PostFormValue("paths"), " ")

    for p:=0; p<len(paths); p++ {
        //err := os.Remove(paths[p])
        err := os.RemoveAll(paths[p])
        if(err != nil) {
            log.Println(err)
            output = err.Error() + "\\n"
        }
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routeLs(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    files, err := ioutil.ReadDir(req.PostFormValue("dir"))
    if(err != nil) {
        log.Println(err)
    }

    json := "{\"files\": ["

    for i, file := range files {
        if(i > 0) {
            json += ", "
        }

        json += "["
        json += strconv.FormatBool(file.IsDir()) + ", "
        json += "\"" + file.Name() + "\", "
        json += "\"" + strconv.Itoa(int(file.Size())) + "\", "
        json += "\"" + file.ModTime().Format("2006-01-02 15:04:05") + "\""
        json += "]"

    }
    json += "]}\n"

    json = strings.ReplaceAll(json, "\\x2d", "\x2d")

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(json))
}

func routePwd(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output, err := os.Getwd()

    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routeTab(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""
    first := true
    files, _ := ioutil.ReadDir(req.PostFormValue("dir"))

    for _, file := range files {
        if(len(file.Name()) >= len(req.PostFormValue("str")) && file.Name()[0:len(req.PostFormValue("str"))] == req.PostFormValue("str")) {
            if(first) {
                first = false
            } else {
                output += " "
            }

            output += file.Name()
        }
    }

    paths := strings.Split(os.Getenv("PATH"), ":")
    programs := make(map[string]bool)

    if(len(req.PostFormValue("str")) > 0) {
        for i:=0; i<len(paths); i++ {
            files, _ = ioutil.ReadDir(paths[i])

            for _, file := range files {
                if(len(file.Name()) >= len(req.PostFormValue("str")) && 
                    file.Name()[0:len(req.PostFormValue("str"))] == req.PostFormValue("str")) {
                    programs[file.Name()] = true;
                }
            }
        }

        for k := range programs {
            if(first) {
                first = false
            } else {
                output += " "
            }

            output += k
        }
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func routePkgMan(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    cmd := "apk --version"
    _, err := exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
    if(err == nil) {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("apk"))
        return
    }

    cmd = "apt -v"
    _, err = exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
    if(err == nil) {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("apt"))
        return
    }

    cmd = "apt-get -v"
    _, err = exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
    if(err == nil) {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("apt-get"))
        return
    }

    cmd = "dnf --version"
    _, err = exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
    if(err == nil) {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("dnf"))
        return
    }

    cmd = "emerge -v"
    _, err = exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
    if(err == nil) {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("emerge"))
        return
    }

    cmd = "pacman -V"
    _, err = exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
    if(err == nil) {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("pacman"))
        return
    }

    cmd = "yum --version"
    _, err = exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
    if(err == nil) {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("yum"))
        return
    }

    cmd = "zypp -version"
    _, err = exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
    if(err == nil) {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("zypp"))
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("[undetected]"))
}

func routeRun(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""
    cmd := req.PostFormValue("command")
    out, err := exec.Command("/bin/bash", "-c", cmd).CombinedOutput()

    if(err != nil) {
        //log.Println(err)
        //output = sanitize(err.Error() + "\n")
    }

    //output = sanitize(string(out))
    output = string(out)

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"output\": \"" + output + "\"}"))
    w.Write([]byte(output))
}

func routeShellRun(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""
    cmd := req.PostFormValue("command")
    shell := req.PostFormValue("shell")
    out, err := exec.Command(shell, "-c", cmd).CombinedOutput()

    if(err != nil) {
        //log.Println(err)
        //output = sanitize(err.Error() + "\n")
    }

    output = string(out)

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(output))
}

func readExecOutput(stdout, stderr *io.ReadCloser, conn *websocket.Conn, ticker *time.Ticker, finished *bool) {
    run_output := ""
    pos_sent := 0
    ticker = time.NewTicker(100 * time.Millisecond)

    go func() {
        for _ = range (*ticker).C {
            var old_pos_sent = pos_sent
            pos_sent = len(run_output)
            if(pos_sent > old_pos_sent) {
                conn.WriteMessage(websocket.TextMessage, []byte(sanitize(run_output[old_pos_sent:pos_sent])))
            }
        }
    }()

    scanner := bufio.NewScanner(io.MultiReader(*stdout, *stderr))
    scanner.Split(bufio.ScanRunes)
    //scanner.Split(bufio.ScanBytes)
    for scanner.Scan() && *finished == false {
        //conn.WriteMessage(websocket.TextMessage, []byte(scanner.Text()))
        run_output += scanner.Text()
    }
    time.Sleep(150 * time.Millisecond)
    ticker.Stop()
    *finished = true
    conn.WriteMessage(websocket.TextMessage, []byte("[finished]"))
}

func routeSocketRun(w http.ResponseWriter, req *http.Request) {
    finished := true
    conn, _ := upgrader.Upgrade(w, req, nil)
    defer conn.Close()

    if(!signatureHandshake(conn)) {
        return
    }

    var stdin io.WriteCloser
    var stdout, stderr io.ReadCloser
    var process *exec.Cmd
    var ticker *time.Ticker

    for true {
        _, cmd, err := conn.ReadMessage()

        if err != nil { //handles page refreshes
            break
        }

        if(string(cmd) == "kill") {
            if(finished == false) {
                finished = true
                process.Process.Kill()
                stdout.Close()
                stderr.Close()
            }
        } else if(!finished) {
            io.WriteString(stdin, string(cmd) + "\n")
        } else if(string(cmd) != "")  {
            finished = false
            process = exec.Command("/bin/bash", "-c", string(cmd))

            stdin, _ = process.StdinPipe()
            stdout, _ = process.StdoutPipe()
            stderr, _ = process.StderrPipe()

            defer stdin.Close()
            defer stdout.Close()
            //defer stderr.Close()

            process.Start()

            go readExecOutput(&stdout, &stderr, conn, ticker, &finished)
        }
    }
}

func readExecOutputPty(ptmx *os.File, conn *websocket.Conn, ticker *time.Ticker, finished *bool) {
    run_output := ""
    pos_sent := 0
    ticker = time.NewTicker(200 * time.Millisecond)

    go func() {
        for _ = range (*ticker).C {
            var old_pos_sent = pos_sent
            pos_sent = len(run_output)
            if(pos_sent > old_pos_sent) {
                conn.WriteMessage(websocket.TextMessage, []byte(sanitize(run_output[old_pos_sent:pos_sent])))
            }
        }
    }()

    scanner := bufio.NewScanner(io.Reader(ptmx))
    scanner.Split(bufio.ScanRunes)
    //scanner.Split(bufio.ScanBytes)
    for scanner.Scan() && *finished == false {
        //conn.WriteMessage(websocket.TextMessage, []byte(scanner.Text()))
        run_output += scanner.Text()
    }
    time.Sleep(250 * time.Millisecond)
    ticker.Stop()
    *finished = true
    conn.WriteMessage(websocket.TextMessage, []byte("[finished]"))
}

func routeSocketRunPty(w http.ResponseWriter, req *http.Request) {
    finished := true
    conn, _ := upgrader.Upgrade(w, req, nil)
    defer conn.Close()

    if(!signatureHandshake(conn)) {
        return
    }

    var process *exec.Cmd
    var ptmx *os.File
    var ticker *time.Ticker

    for true {
        _, cmd, err := conn.ReadMessage()

        if err != nil { //handles page refreshes
            break
        }

        if(string(cmd) == "kill") {
            if(finished == false) {
                finished = true
                process.Process.Kill()
                ptmx.Close()
                //stdout.Close()
                //stderr.Close()
            }
        } else if(!finished) {
            ptmx.Write([]byte(string(cmd) + "\n"))
        } else if(string(cmd) != "")  {
            finished = false
            process = exec.Command("/bin/bash", "-c", string(cmd))

            ptmx, err = pty.Start(process)
            if err != nil {
                log.Println(err)
                continue
            }  
            defer ptmx.Close()

            go readExecOutputPty(ptmx, conn, ticker, &finished)
        }
    }
}

func routeDownload(w http.ResponseWriter, req *http.Request) {
    //fmt.Println(req.URL.String())
    //fmt.Println(req.PostFormValue("path"))

    //w.Header().Set("Content-Disposition", "attachment; filename=" + req.PostFormValue("path"))
    //w.Header().Set("Content-Type", "application/octet-stream")
    http.ServeFile(w, req, req.PostFormValue("path"))
}


func routeUpload(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    file, _, err := req.FormFile("file")
    if err != nil {
        w.Header().Set("Content-Type", "text/plain")
        w.Write([]byte("{\"output\": \"error retrieving file\"}"))
        return
    }
    defer file.Close()
    f, err := os.OpenFile(req.PostFormValue("path"), os.O_WRONLY|os.O_CREATE, 0666)
    defer f.Close()
    io.Copy(f, file)

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("upload successful"))
}

func routeVerifySignature(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("true"))
}

func createConfig() {
    if(!fileExists(home + "/.cloudfort/")) {
        err := os.Mkdir(home + "/.cloudfort/", 0755)
        if(err != nil) {
            log.Fatal(err)
        } 
    }

    err := os.WriteFile(home + "/.cloudfort/config.json", []byte(
`{
    "app": {
    },

    "editor": {
        "highlight": true
    },

    "monitor": {
        "sort-by": "memory"
    },

    "terminal": {
        "mode": "pty"
    }
}`), 0644);

    if(err != nil) {
        log.Fatal(err)
    }
}

func createReposConfig() {
    if(!fileExists(home + "/.cloudfort/")) {
        err := os.Mkdir(home + "/.cloudfort/", 0755)
        if(err != nil) {
            log.Fatal(err)
        } 
    }

    err := os.WriteFile(home + "/.cloudfort/repositories.json", []byte(
`{
    "repositories": {
    }
}`), 0644);

    if(err != nil) {
        log.Fatal(err)
    }
}

func main() {
    cmd := os.Args[1]

    if(cmd == "serve") {
        mux := http.NewServeMux()
        mux.HandleFunc("/api/cd",          routeCd)
        mux.HandleFunc("/api/deploy",      routeDeploy)
        mux.HandleFunc("/api/dir/create",  routeDirCreate)
        mux.HandleFunc("/api/file/create", routeFileCreate)
        mux.HandleFunc("/api/file/read",   routeFileRead)
        mux.HandleFunc("/api/file/write",  routeFileWrite)
        mux.HandleFunc("/api/home",        routeHome)
        mux.HandleFunc("/api/rename",      routeRename)
        mux.HandleFunc("/api/mv",          routeMv)
        mux.HandleFunc("/api/cp",          routeCp)
        mux.HandleFunc("/api/rm",          routeRm)
        mux.HandleFunc("/api/ls",          routeLs)
        mux.HandleFunc("/api/pwd",         routePwd)
        mux.HandleFunc("/api/tab",         routeTab)
        mux.HandleFunc("/api/pkg-man",     routePkgMan)
        mux.HandleFunc("/api/run",         routeRun)
        mux.HandleFunc("/api/shell/run",   routeShellRun)
        mux.HandleFunc("/api/socket/run",  routeSocketRun)
        mux.HandleFunc("/api/socket/pty",  routeSocketRunPty)
        mux.HandleFunc("/api/download",    routeDownload)
        mux.HandleFunc("/api/upload",      routeUpload)
        mux.HandleFunc("/api/sigcheck",    routeVerifySignature)

        if(os.Args[2] != "http" && os.Args[2] != "https") {
            fmt.Println("second parameter should be 'http' or 'https'")
            return
        }

        https := (os.Args[2] == "https")
        domain = os.Args[3]
        port = os.Args[4]
        password = []byte(os.Args[5])

        home, _ = os.UserHomeDir()

        if(!fileExists(home + "/.cloudfort/config.json")) {
            createConfig()
        }

        if(!fileExists(home + "/.cloudfort/repositories.json")) {
            createReposConfig()
        }

        config_bytes, _ := os.ReadFile(home + "/.cloudfort/config.json");
        config, _ = gabs.ParseJSON(config_bytes)

        repos_bytes, _ := os.ReadFile(home + "/.cloudfort/repositories.json");
        repos, _ = gabs.ParseJSON(repos_bytes)

        //fmt.Println(config); 
        //fmt.Println(repos); 
        
        pwd, _ := os.Getwd()
        if(pwd == "/") {
            os.Chdir("/root/")
        }
        
        mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("/var/www/cloudfort-dash/public"))))

        if(https) {
            fmt.Println("Serving Cloudfort Dash over https on port " + port)
            for {
                err := http.ListenAndServeTLS(":" + port, 
                    "/etc/letsencrypt/live/" + domain + "/fullchain.pem", 
                    "/etc/letsencrypt/live/" + domain + "/privkey.pem", 
                    mux);
                if err != nil {
                    log.Println("ListenAndServeTLS: ", err)
                }
            }
        } else {
            fmt.Println("Serving Cloudfort Dash over http on port " + port)
            for {
                err := http.ListenAndServe(":" + port, mux)
                if err != nil {
                    log.Println("ListenAndServe: ", err)
                }
            }
        } 
    } else if(cmd == "update" || cmd == "upgrade" || cmd == "force-update") {
        cmd := "curl -s https://raw.githubusercontent.com/cloudfort-app/cloudfort-dash/main/version.md"
        latestVersion, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()

        if(err != nil) {
            log.Println(err)
            return
        }

        if(cmd != "force-update" && string(latestVersion) == version) {
            fmt.Println("latest version already installed")
            return
        }

        cmd = "wget -q -O cloudfort-dash.tar.xz https://github.com/cloudfort-app/cloudfort-dash/releases/download/" + string(latestVersion) + "/cloudfort-dash.tar.xz; "
        cmd += "tar -xvf cloudfort-dash.tar.xz; "
        cmd += "rm cloudfort-dash.tar.xz; "
        cmd += "rm -r /var/www/cloudfort-dash/public; "
        cmd += "mv public /var/www/cloudfort-dash; "
        cmd += "chmod a+x cloudfort-dash; "
        cmd += "rm /usr/local/bin/cloudfort-dash; "
        cmd += "mv cloudfort-dash /usr/local/bin; "
        out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()

        if(err != nil) {
            log.Println(err)
            return
        }

        fmt.Println(string(out))
        fmt.Println("cloudfort-dash updated successfully")

    } else if(cmd == "--version") {
        fmt.Println(version)
    } else {
        fmt.Println("do not recognize command '" + cmd + "'")
    }
}
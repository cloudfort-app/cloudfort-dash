package main

import (
    //"bytes"
    "encoding/csv"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
)

var no_started = 0
var no_finished = 0
var max_animating = 10

func getCd(w http.ResponseWriter, req *http.Request) {
    output := ""

    err := os.Chdir(req.PostFormValue("dir"))
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"output\": \"" + output + "\"}"))
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
    // io.WriteString(w, "hello, world!\n")
}

func postCreateDir(w http.ResponseWriter, req *http.Request) {
    output := ""

    err := os.Mkdir(req.PostFormValue("path"), 0755)
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    } 

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"output\": \"" + output + "\"}"))
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
    // io.WriteString(w, "hello, world!\n")
}

func postCreateFile(w http.ResponseWriter, req *http.Request) {
    output := ""

    file, err := os.Create(req.PostFormValue("path"))
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    } else {
        defer file.Close()
    }

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"output\": \"" + output + "\"}"))
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
    //fmt.Fprintf(w, "{}")
    // io.WriteString(w, "hello, world!\n")
}

func postMv(w http.ResponseWriter, req *http.Request) {
    output := ""

    err := os.Rename(req.PostFormValue("path-old"), req.PostFormValue("path-new"))
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"output\": \"" + output + "\"}"))
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
    //fmt.Fprintf(w, "{}")
    // io.WriteString(w, "hello, world!\n")
}

func postRm(w http.ResponseWriter, req *http.Request) {
    output := ""

    err := os.Remove(req.PostFormValue("path"))
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"output\": \"" + output + "\"}"))
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
    //fmt.Fprintf(w, "{}")
    // io.WriteString(w, "hello, world!\n")
}

func sanitize(in_str string) string {
    str := strings.Replace(in_str, "\r\n", "\n", -1)

    m1 := regexp.MustCompile(`\n.*\r`)
    str = m1.ReplaceAllString(str, "\n")

    str = strings.Replace(str, "\n", "\\n", -1)
    str = strings.Replace(str, "\t", "\\t", -1)

    str = strings.Replace(str, "\\\"", "\"", -1)
    str = strings.Replace(str, "\"", "\\\"", -1)

    return str;
}

func getFileStr(w http.ResponseWriter, req *http.Request) {
    b, err := os.ReadFile(req.PostFormValue("path")) 
    if err != nil {
        log.Println(err)
    }

    str := sanitize(string(b));

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"content\": \"" + str + "\"}"))
    fmt.Fprintf(w, "{\"content\": \"" + str + "\"}")
}

func postWriteFile(w http.ResponseWriter, req *http.Request) {
    output := ""

    err := os.WriteFile(req.PostFormValue("path"), []byte(req.PostFormValue("content")), 0644);
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"output\": \"" + output + "\"}"))
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
    //fmt.Fprintf(w, "{}")
    // io.WriteString(w, "hello, world!\n")
}

func getLs(w http.ResponseWriter, req *http.Request) {
    files, err := ioutil.ReadDir(req.PostFormValue("dir"))
    if(err != nil) {
        log.Println(err);
    }

    json := "{\"files\": [";

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
    json += "]}\n";

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte(json))
    // fmt.Fprintf(w, "hello, world!\n")
    // io.WriteString(w, "hello, world!\n")
}

func getPwd(w http.ResponseWriter, req *http.Request) {
    output, err := os.Getwd()

    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("{\"pwd\": \"" + output + "\"}"))
    // fmt.Fprintf(w, "hello, world!\n")
    // io.WriteString(w, "hello, world!\n")
}

func split_params(params_str string) []string {
    r := csv.NewReader(strings.NewReader(params_str))
    r.Comma = ' '

    record, err := r.Read()
    if err != nil {
        log.Println(err)
    }

    return record
}

func getRun(w http.ResponseWriter, req *http.Request) {
    output := ""
    cmd := req.PostFormValue("command")
    out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()

    if(err != nil) {
        log.Println(err)
        output = sanitize(err.Error() + "\n")
    } else {
        output = sanitize(string(out));
    }

    w.Header().Set("Content-Type", "text/plain")
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
}

/*func getRunOrig(w http.ResponseWriter, req *http.Request) {
    output := ""
    cmd := req.PostFormValue("command")
    params := split_params(req.PostFormValue("params"))

    var out []byte
    var err error

    if(req.PostFormValue("params") == "") {
        out, err = exec.Command(cmd).CombinedOutput()
    } else {
        out, err = exec.Command(cmd, params...).CombinedOutput()
    }

    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
        output = sanitize(output);
    } else {
        output = sanitize(string(out));
        fmt.Println(output)
    }

    w.Header().Set("Content-Type", "text/plain")
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
}*/

/*impot "bytes"
func getRunWithPipes(w http.ResponseWriter, req *http.Request) {
    output := ""
    cmds := strings.Split(req.PostFormValue("command"), " | ")
    var params [][]string
    var execs []*exec.Cmd
    var err error

    for i:=0; i<len(cmds); i++ {
        params = append(params, split_params(cmds[i]))
        cmds[i] = params[i][0]
        params[i] = params[i][1:]

        if(len(params[i]) == 0) {
            execs = append(execs, exec.Command(cmds[i]))
        } else {
            execs = append(execs, exec.Command(cmds[i], params[i]...))
        }

        if(i > 0) {
            execs[i].Stdin, err = execs[i-1].StdoutPipe()
            if(err != nil) {
                log.Println(err)
                output = sanitize(err.Error() + "\n")

                w.Header().Set("Content-Type", "text/plain")
                fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
                return;
            }
        }
    }

    var b bytes.Buffer
    execs[len(cmds)-1].Stdout = &b

    for i:=len(cmds)-1; i>0; i-- {
        err = execs[i].Start()
        if(err != nil) {
            log.Println(err)
            output = sanitize(err.Error() + "\n")

            w.Header().Set("Content-Type", "text/plain")
            fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
            return;
        }
    }

    err = execs[0].Run()
    if(err != nil) {
        log.Println(err)
        output = sanitize(err.Error() + "\n")

        w.Header().Set("Content-Type", "text/plain")
        fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
        return;
    }

    for i:=1; i<len(cmds); i++ {
        err = execs[i].Wait()
        if(err != nil) {
            log.Println(err)
            output = sanitize(err.Error() + "\n")

            w.Header().Set("Content-Type", "text/plain")
            fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
            return;
        }
    }

    output = sanitize(b.String())

    w.Header().Set("Content-Type", "text/plain")
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
}*/

func getTop(w http.ResponseWriter, req *http.Request) {
    output := ""

    out, err := exec.Command("top", "-bn1").CombinedOutput()

    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
        output = sanitize(output);
    } else {
        output = sanitize(string(out));
    }

    w.Header().Set("Content-Type", "text/plain")
    //w.Write([]byte("{\"output\": \"" + output + "\"}"))
    fmt.Fprintf(w, "{\"output\": \"" + output + "\"}")
    // io.WriteString(w, "hello, world!\n")
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/cd", getCd)
    mux.HandleFunc("/createdir", postCreateDir)
    mux.HandleFunc("/createfile", postCreateFile)
    mux.HandleFunc("/mv", postMv)
    mux.HandleFunc("/rm", postRm)
    mux.HandleFunc("/writefile", postWriteFile)
    mux.HandleFunc("/filestr", getFileStr)
    mux.HandleFunc("/ls", getLs)
    mux.HandleFunc("/pwd", getPwd)
    mux.HandleFunc("/run", getRun)
    mux.HandleFunc("/top", getTop)
    mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./public"))))
    //http.HandleFunc("/hello", getHello)

    fmt.Println("hosting http-server on port 80")
    err := http.ListenAndServe(":80", mux)
    if err != nil {
        log.Println("Listen: ", err)
    }
    
    /*fmt.Println("hosting http-server on port 443")
    err := http.ListenAndServeTLS(":443", 
        "/etc/letsencrypt/live/granimate.art/fullchain.pem", 
        "/etc/letsencrypt/live/granimate.art/privkey.pem", 
        mux);
    if err != nil {
        log.Println("ListenAndServe: ", err)
    }*/
}
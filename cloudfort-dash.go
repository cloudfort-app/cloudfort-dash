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

//var domain = "https://dash.cloudfort.app"
//var domain = "http://74.207.230.17:8000/"
//var domain = "http://localhost:8000/"
var domain = os.Args[1]
var port = os.Args[2]

/*func check_referer(req *http.Request) bool {
    return (req.Referer() == "" || req.Referer()[0:len(domain)] != domain)
}*/

func getCd(w http.ResponseWriter, req *http.Request) {
    /*if(check_referer(req)) {
        w.Write([]byte("404 page not found"))
        return
    }*/

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

    str = strings.Replace(str, "\\", "\\\\", -1)

    str = strings.Replace(str, "\n", "\\n", -1)
    str = strings.Replace(str, "\t", "\\t", -1)
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
    mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./dash"))))
    //http.HandleFunc("/hello", getHello)

    if(port == "80" || port == "8000") {
        fmt.Println("hosting Cloudfort Dash on port " + port)
        err := http.ListenAndServe(":" + port, mux)
        if err != nil {
            log.Println("Listen: ", err)
        }
    } else {
        fmt.Println("hosting Cloudfort Dash on port " + port)
        err := http.ListenAndServeTLS(":" + port, 
            "/etc/letsencrypt/live/" + domain + "/fullchain.pem", 
            "/etc/letsencrypt/live/" + domain + "/privkey.pem", 
            mux);
        if err != nil {
            log.Println("ListenAndServe: ", err)
        }
    }
}
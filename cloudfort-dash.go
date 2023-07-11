package main

import (
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
)

var config *gabs.Container
var domain string
var home string
var port string
var password []byte
var version = "v0.1.6"

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

func hasValidSignature(w *http.ResponseWriter, req *http.Request) bool {
    mac := hmac.New(sha256.New, password)
    mac.Write([]byte(req.PostFormValue("date")))

    req_date, _ := strconv.Atoi(req.PostFormValue("date"))

    if(req_date > int(time.Now().UnixMilli())) {
        fmt.Println("request is from the future")
        (*w).Header().Set("Content-Type", "text/plain")
        (*w).Write([]byte("{\"msg\": \"invalid signature\"}"))

        return false;
    } else if(int(time.Now().UnixMilli()) - req_date > 5000) {
        fmt.Println("request is older than 5 seconds")
        (*w).Header().Set("Content-Type", "text/plain")
        (*w).Write([]byte("{\"msg\": \"invalid signature\"}"))

        return false;
    } else if(hex.EncodeToString(mac.Sum(nil)) != req.Header.Get("signature")) {
        fmt.Println("request has invalid signature")
        (*w).Header().Set("Content-Type", "text/plain")
        (*w).Write([]byte("{\"msg\": \"invalid signature\"}"))

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

    str = strings.Replace(str, "\n", "\\n", -1)
    str = strings.Replace(str, "\t", "\\t", -1)

    str = strings.Replace(str, "\b", "\\b", -1)
    str = strings.Replace(str, "\f", "\\f", -1)
    str = strings.Replace(str, "\r", "\\r", -1)

    str = strings.Replace(str, "\"", "\\\"", -1)

    return str;
}

func sanitize_for_saving(in_str string) string {
    str := strings.Replace(in_str, "\r\n", "\n", -1)

    return str;
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
    w.Write([]byte("{\"output\": \"" + output + "\"}"))
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
    w.Write([]byte("{\"output\": \"" + output + "\"}"))
}

func routeDeploy(w http.ResponseWriter, req *http.Request) {
    fmt.Println("received deploy request");

    body_bytes, _ := io.ReadAll(req.Body);

    if(!hasValidWebhookSignature(&w, req, body_bytes)) {
        return
    }
    
    body, _ := gabs.ParseJSON(body_bytes)
    name, _ := body.Path("repository.full_name").Data().(string)
    deploy_cmd, _ := config.Path("repositories." + name + ".deploy").Data().(string)
    out, err := exec.Command("/bin/sh", "-c", deploy_cmd).CombinedOutput()

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
        output = err.Error() + "\\n"
    } else {
        defer file.Close()
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("{\"output\": \"" + output + "\"}"))
}

func routeFileRead(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    b, err := os.ReadFile(req.PostFormValue("path")) 
    if err != nil {
        log.Println(err)
    }

    str := sanitize(string(b));

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("{\"content\": \"" + str + "\"}"))
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
    w.Write([]byte("{\"output\": \"" + output + "\"}"))
}

func routeHome(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("{\"home\": \"" + home + "\"}"))
}

func routeMv(w http.ResponseWriter, req *http.Request) {
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
    w.Write([]byte("{\"output\": \"" + output + "\"}"))
}

func routeRm(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""

    err := os.Remove(req.PostFormValue("path"))
    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("{\"output\": \"" + output + "\"}"))
}

func routeLs(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

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
    w.Write([]byte("{\"pwd\": \"" + output + "\"}"))
}

func routeRun(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""
    cmd := req.PostFormValue("command")
    out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()

    if(err != nil) {
        log.Println(err)
        output = sanitize(err.Error() + "\n")
    } else {
        output = sanitize(string(out))
        //fmt.Println("{\"output\": \"" + output + "\"}")
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("{\"output\": \"" + output + "\"}"))
}

func routePasswordCheck(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("{\"valid\": true}"))
}

func routeTop(w http.ResponseWriter, req *http.Request) {
    if(!hasValidSignature(&w, req)) {
        return
    }

    output := ""

    out, err := exec.Command("top", "-bn1", "-o", "%MEM", "-w", "120").CombinedOutput()

    if(err != nil) {
        log.Println(err)
        output = err.Error() + "\\n"
        output = sanitize(output);
    } else {
        output = sanitize(string(out));
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("{\"output\": \"" + output + "\"}"))
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
    "repositories": {}
}`), 0644);

    if(err != nil) {
        log.Fatal(err)
    }
}

func main() {
    cmd := os.Args[1]

    if(cmd == "serve") {
        mux := http.NewServeMux()
        mux.HandleFunc("/api/cd",           routeCd)
        mux.HandleFunc("/api/deploy",       routeDeploy)
        mux.HandleFunc("/api/dir/create",   routeDirCreate)
        mux.HandleFunc("/api/file/create",  routeFileCreate)
        mux.HandleFunc("/api/file/read",    routeFileRead)
        mux.HandleFunc("/api/file/write",   routeFileWrite)
        mux.HandleFunc("/api/home",         routeHome)
        mux.HandleFunc("/api/mv",           routeMv)
        mux.HandleFunc("/api/rm",           routeRm)
        mux.HandleFunc("/api/ls",           routeLs)
        mux.HandleFunc("/api/pwd",          routePwd)
        mux.HandleFunc("/api/run",          routeRun)
        mux.HandleFunc("/api/password/check", routePasswordCheck)
        mux.HandleFunc("/api/top",          routeTop)

        domain = os.Args[2]
        port = os.Args[3]
        password = []byte(os.Args[4])

        home, _ = os.UserHomeDir()

        if(!fileExists(home + "/.cloudfort/config.json")) {
            createConfig()
        }

        config_bytes, _ := os.ReadFile(home + "/.cloudfort/config.json");
        config, _ = gabs.ParseJSON(config_bytes)

        //fmt.Println(config); 
        
        pwd, _ := os.Getwd()
        if(pwd == "/") {
            os.Chdir("/root/")
        }
        
        mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("/var/www/cloudfort-dash/public"))))

        if(port == "80" || port == "8000" || port == "8080") {
            fmt.Println("Hosting Cloudfort Dash (http) on port " + port)
            err := http.ListenAndServe(":" + port, mux)
            if err != nil {
                log.Println("ListenAndServe: ", err)
            }
        } else {
            fmt.Println("Hosting Cloudfort Dash (https) on port " + port)
            err := http.ListenAndServeTLS(":" + port, 
                "/etc/letsencrypt/live/" + domain + "/fullchain.pem", 
                "/etc/letsencrypt/live/" + domain + "/privkey.pem", 
                mux);
            if err != nil {
                log.Println("ListenAndServeTLS: ", err)
            }
        }
    } else if(cmd == "update") {
        cmd := "curl -s https://raw.githubusercontent.com/cloudfort-app/cloudfort-dash/main/version.md"
        latestVersion, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()

        if(err != nil) {
            log.Println(err)
            return
        }

        if(string(latestVersion) == version) {
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
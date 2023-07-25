# Cloudfort Dash ⚡

[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](/LICENSE)
[![source code](https://img.shields.io/github/languages/code-size/cloudfort-app/cloudfort-dash.svg?label=Source%20Code)](https://github.com/cloudfort-app/cloudfort-dash)
[![language](https://img.shields.io/badge/Language-Go-blue)](https://cloudfort.app/products/dash/)

\[[website](https://cloudfort.app/products/dash/)\]
\[[repository](https://github.com/cloudfort-app/cloudfort-dash)\]
\[[documentation](https://cloudfort.app/docs/dash/)\]

> 🛠 Dash provides an os-like experience in the browser for unix based servers. 

## Features
* Terminal
	- Interactive
	- Tab Completion
* File Explorer
	- Browse File System
	- Create/Delete/Copy/Move/Edit Files
	- Compress/Extract
	- Upload/Download
	- View Images/Videos
* Editor
	- Syntax Highlighting
	- Line Numbers
	- Find and Replace
	- Regex Replace
* Monitor
* Tools
	- Certs
	- Cronjobs
	- Docker
	- fail2ban
	- Firewall
	- Git Repositories
	- Services
	- SSH
	- User Management
	- Websites 

## Install
```
wget https://cloudfort.app/dash/install.sh && chmod a+x install.sh && ./install.sh
```

## Serving
```
cloudfort-dash serve http(s) domain port secret
```

Examples:
```
cloudfort-dash serve http 194.195.248.93 8080 "the cat sat on the mat"
cloudfort-dash serve https example.com 4443 "the cat sat on the mat"
cloudfort-dash serve http localhost 80 "the cat sat on the mat"
```

Use example 1 for http access, example 2 for https access and example 3 for a tunneling solution like ngrok.

## Browser 
Examples:
```
http://194.195.248.93:8080/
https://examples.com:4443/
```

## Example Service File
```
[Unit] 
Description = cloudfort-dash

[Service]
Type           = simple
User           = root
Group          = root
LimitNOFILE    = 4096
Restart        = always
RestartSec     = 5s
StandardOutput = append:/usr/local/bin/cloudfort-dash.errors.log
StandardError  = append:/usr/local/bin/cloudfort-dash.errors.log
ExecStart      = /usr/local/bin/cloudfort-dash 172.167.201.130 8080 "this is a secret"

[Install]
WantedBy = multi-user.target
```

Copyright (c) 2023-present [Nicholas Ham](https://n-ham.com).

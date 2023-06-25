#!/usr/bin/env bash

echo "[Unit]
Description = pocketbase

[Service]
Type           = simple
User           = root
Group          = root
LimitNOFILE    = 4096
Restart        = always
RestartSec     = 5s
StandardOutput = append:/usr/local/bin/pocketbase.errors.log
StandardError  = append:/usr/local/bin/pocketbase.errors.log
ExecStart      = /usr/local/bin/pocketbase serve --http=\"$1:8000\"

[Install]
WantedBy = multi-user.target" > pocketbase.service

echo "[Unit] 
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
ExecStart      = /usr/local/bin/cloudfort-dash $1 8080

[Install]
WantedBy = multi-user.target" > cloudfort-dash.service

echo "# This file controls the state of SELinux on the system.
# SELINUX= can take one of these three values:
#       enforcing - SELinux security policy is enforced.
#       permissive - SELinux prints warnings instead of enforcing.
#       disabled - No SELinux policy is loaded.
SELINUX=permissive
# SELINUXTYPE= can take one of these two values:
#       targeted - Targeted processes are protected,
#       mls - Multi Level Security protection.
SELINUXTYPE=targeted" > /etc/selinux/config

mv pocketbase.service cloudfort-dash.service /etc/systemd/system/
setenforce 0
systemctl daemon-reload;
systemctl enable pocketbase.service;
systemctl enable cloudfort-dash.service;
systemctl start pocketbase;
systemctl start cloudfort-dash; 
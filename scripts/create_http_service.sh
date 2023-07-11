#!/usr/bin/env bash

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
ExecStart      = /usr/local/bin/cloudfort-dash serve $1 8080 \"$2\"

[Install]
WantedBy = multi-user.target" > /etc/systemd/system/cloudfort-dash.service

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

setenforce 0
systemctl daemon-reload;
systemctl enable cloudfort-dash.service;
systemctl start cloudfort-dash; 
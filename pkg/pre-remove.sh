#!/bin/bash

BIN_DIR=/usr/bin

# Distribution-specific logic
if [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        deb-systemd-invoke stop wavefront-proxy.service
    else
        # Assuming sysv
        invoke-rc.d wavefront-proxy stop
    fi
fi

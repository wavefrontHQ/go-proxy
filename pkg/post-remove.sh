#!/bin/bash

function disable_systemd {
    systemctl disable wavefront-proxy
    rm -f $1
}

function disable_update_rcd {
    update-rc.d -f wavefront-proxy remove
    rm -f /etc/init.d/wavefront-proxy
}

function disable_chkconfig {
    chkconfig --del wavefront-proxy
    rm -f /etc/init.d/wavefront-proxy
}

if [[ -f /etc/redhat-release ]] || [[ -f /etc/SuSE-release ]]; then
    # RHEL-variant logic
    if [[ "$1" = "0" ]]; then
        rm -f /etc/default/wavefront-proxy

        if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
            disable_systemd /usr/lib/systemd/system/wavefront-proxy.service
        else
            # Assuming sysv
            disable_chkconfig
        fi
    fi
elif [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic
    if [ "$1" == "remove" -o "$1" == "purge" ]; then
        # Remove/purge
        rm -f /etc/default/wavefront-proxy

        if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
            disable_systemd /lib/systemd/system/wavefront-proxy.service
        else
            # Assuming sysv
            # Run update-rc.d or fallback to chkconfig if not available
            if which update-rc.d &>/dev/null; then
                disable_update_rcd
            else
                disable_chkconfig
            fi
        fi
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ $ID = "amzn" ]]; then
        # Amazon Linux logic
        if [[ "$1" = "0" ]]; then
            rm -f /etc/default/wavefront-proxy
            disable_chkconfig
        fi
    fi
fi

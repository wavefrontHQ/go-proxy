#!/bin/bash

USER=wavefront
GROUP=wavefront
BIN_DIR=/usr/bin
LOG_DIR=/var/log/wavefront
SCRIPT_DIR=/usr/lib/wavefront-proxy/scripts
WKG_DIR=/etc/wavefront/wavefront-proxy
LOGROTATE_DIR=/etc/logrotate.d

function install_init {
    cp -f $SCRIPT_DIR/init.sh /etc/init.d/wavfront-proxy
    chmod +x /etc/init.d/wavfront-proxy
}

function install_systemd {
    cp -f $SCRIPT_DIR/wavefront-proxy.service $1
    systemctl enable wavefront-proxy || true
    systemctl daemon-reload || true
}

function install_update_rcd {
    update-rc.d wavefront-proxy defaults
}

function install_chkconfig {
    chkconfig --add wavefront-proxy
}

if ! grep "^${GROUP}:" /etc/group &>/dev/null; then
    groupadd -r ${GROUP}
fi

if ! id ${USER} &>/dev/null; then
    useradd -r -s /bin/bash -g ${GROUP} ${USER} &> /dev/null
fi

test -d $LOG_DIR || mkdir -p $LOG_DIR
chown -R -L ${USER}:${GROUP} $LOG_DIR
chmod 755 $LOG_DIR

chown -R -L ${USER}:${GROUP} $WKG_DIR
chmod 755 $WKG_DIR

# Add defaults file, if it doesn't exist
if [[ ! -f /etc/default/wavefront-proxy ]]; then
    touch /etc/default/wavefront-proxy
fi

if [[ ! -f ${WKG_DIR}/wavefront.conf ]] ; then
    cp ${WKG_DIR}/wavefront.conf.default ${WKG_DIR}/wavefront.conf
    chown ${USER}:${GROUP} ${WKG_DIR}/wavefront.conf
fi

# Distribution-specific logic
if [[ -f /etc/redhat-release ]] || [[ -f /etc/SuSE-release ]]; then
    # RHEL-variant logic
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        install_systemd /usr/lib/systemd/system/wavefront-proxy.service
    else
        # Assuming SysVinit
        install_init
        # Run update-rc.d or fallback to chkconfig if not available
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
    fi
elif [[ -f /etc/debian_version ]]; then
    # Debian/Ubuntu logic
    if [[ "$(readlink /proc/1/exe)" == */systemd ]]; then
        install_systemd /lib/systemd/system/wavefront-proxy.service
        systemctl restart wavefront-proxy || echo "WARNING: systemd not running."
    else
        # Assuming SysVinit
        install_init
        # Run update-rc.d or fallback to chkconfig if not available
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
        invoke-rc.d wavefront-proxy restart
    fi
elif [[ -f /etc/os-release ]]; then
    source /etc/os-release
    if [[ $ID = "amzn" ]]; then
        # Amazon Linux logic
        #TODO: uncomment this after initd is fixed
        install_init
        # Run update-rc.d or fallback to chkconfig if not available
        if which update-rc.d &>/dev/null; then
            install_update_rcd
        else
            install_chkconfig
        fi
    fi
fi

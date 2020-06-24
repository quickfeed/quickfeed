#!/bin/bash

if [ $# -ne 1 ] && [ $# -ne 3 ]; then
    echo -e "usage: $0 remote_user [remote_host] [local_port]";
    exit;
fi

username=$1
rhost="ag3.ux.uis.no"
lport="8081"

if [[ $# -eq 3 ]]; then
    rhost=$2
    lport=$3
fi

case $username in
    ag2)
        rport=3001
        ;;
    meling)
        rport=3002
        ;;
    nicolasf)
        rport=3003
        ;;
    veray)
        rport=3004
        ;;
    frtvedt)
        rport=3006
        ;;
      *)
        echo -e "$0: invalid username: $username";
        echo -e "usage: $0 remote_user remote_host local_port";
        exit;
esac

gnome-terminal -e "ssh -v -L 7575:$rhost:22 -N $username@badne5.ux.uis.no"
sleep 5
ssh -v -R $rport:127.0.0.1:$lport -N $username@localhost -p 7575

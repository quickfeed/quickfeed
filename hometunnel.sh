#!/bin/bash

if [ $# -ne 1 ] && [ $# -ne 3 ]; then
    echo -e "usage: $0 remote_user [remote_host] [local_port]";
    exit;
fi

username=$1
rhost="ag2.ux.uis.no"
lport="8080"

if [[ $# -eq 3 ]]; then
    rhost=$2
    lport=$3
fi

case $username in
    pedersen)
        rport=3001
        ;;
    meling)
        rport=3002
        ;;
    nicolasf)
        rport=3003
        ;;
    junaid)
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

gnome-terminal -e "ssh -L 7575:$rhost:22 -N $username@badne7.ux.uis.no"
sleep 5
ssh -R $rport:localhost:$lport -N $username@localhost -p 7575
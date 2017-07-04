#!/bin/bash

if [[ $# -ne 3 ]]; then
    echo -e "usage: $0 remote_user remote_host local_port";
    exit;
fi

username=$1
rhost=$2
lport=$3

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
    *)
        echo -e "$0: invalid username: $username";
        echo -e "usage: $0 remote_user remote_host local_port";
        exit;
esac

ssh -R $rport:localhost:$lport -N $username@$rhost

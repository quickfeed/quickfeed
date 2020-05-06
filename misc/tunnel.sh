#!/bin/bash

if [ $# -ne 1 ] && [ $# -ne 3 ]; then
    echo -e "usage: $0 remote_user [remote_host] [local_port]";
    exit;
fi

username=$1
rhost="ag2.ux.uis.no"
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

ssh -v -R $rport:localhost:$lport -N $username@$rhost

# ssh -v -R 9090:localhost:9090 -N veray@ag2.ux.uis.no
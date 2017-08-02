#!/bin/bash

if [ $# -ne 2 ]; then
    echo "usage: $0 username baseurl"
    echo ""
    echo "This command will start a gradle container with the script testjava"
    echo "The username is the repository it should pull from, and base url is"
    echo "organisation url. This has to support baseurl/[username]-assignments"
    echo "and baseurl/tests, to work properly"
    exit 0
fi

JAVATEST=testjava.sh
    
sudo docker run --rm -v $PWD/$JAVATEST:/$JAVATEST gradle bash /$JAVATEST $1 $2

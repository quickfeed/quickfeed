#!/bin/bash
if [ ! -d "/etc/dummycerts" ]
then
	sudo mkdir /etc/dummycerts
	echo "Creating dummy Certificate folder"
fi
#Folder for 
cd /etc/dummycerts/

#Creating certificates
sudo openssl req -x509 -nodes -new -sha256 -days 1024 -newkey rsa:2048 -keyout RootCA.key -out RootCA.pem -subj "/C=US/CN=Example-Root-CA"
sudo openssl x509 -outform pem -in RootCA.pem -out RootCA.crt


echo -e  "\e[31mWarning\e[39m: These certificates are for development only and should never be used in production"
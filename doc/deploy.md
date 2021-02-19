# Deploy

## Deploy on Ubuntu

Here is commands to upgrade `certbot` according to [Certbot instructions](https://certbot.eff.org/lets-encrypt/ubuntufocal-nginx).

```sh
apt-get remove certbot
snap install core
snap refresh core
snap install --classic certbot
ln -s /snap/bin/certbot /usr/bin/certbot
```

## Deploy on macOS

Install tools.

```sh
brew install certbot
brew install nginx
```

### Instructions from nginx

```text
Docroot is: /usr/local/var/www

The default port has been set in /usr/local/etc/nginx/nginx.conf to 8080 so that
nginx can run without sudo.

nginx will load all files in /usr/local/etc/nginx/servers/.

To have launchd start nginx now and restart at login:
  brew services start nginx
Or, if you don't want/need a background service you can just run:
  nginx
```

I just ran `nginx` directly for testing.

### Configure Fixed IP and Router

Configure NameCheap with your fixed IP; in my case 92.221.105.172.

```text
Type         Host          Value                TTL
A Record     cyclone       92.221.105.172       5 min
```

Set up port forwarding on Google WiFi for Heins-Cyclone.
TCP.
External ports 80.
Internal ports 8080.

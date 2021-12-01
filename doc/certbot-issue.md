# Certbot Manual Renew

Envoy does not seem to support auto renew of certificates via certbot.
See [here for details](https://github.com/envoyproxy/envoy/issues/96).

```shell
% sudo certbot renew --standalone
Saving debug log to /var/log/letsencrypt/letsencrypt.log

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Processing /etc/letsencrypt/renewal/itest.run.conf
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Error while running nginx -c /etc/nginx/nginx.conf -t.

nginx: [emerg] cannot load certificate "/etc/letsencrypt/live/itest.run-0001/fullchain.pem": BIO_new_file() failed (SSL: error:02001002:system library:fopen:No such file or directory:fopen('/etc/letsencrypt/live/itest.run-0001/fullchain.pem','r') error:2006D080:BIO routines:BIO_new_file:no such file)
nginx: configuration file /etc/nginx/nginx.conf test failed

Renewing an existing certificate for itest.run and 3 more domains

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Congratulations, all renewals succeeded:
  /etc/letsencrypt/live/itest.run/fullchain.pem (success)
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

% sudo certbot certificates
Saving debug log to /var/log/letsencrypt/letsencrypt.log

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
Found the following certs:
  Certificate Name: itest.run
    Serial Number: 415a1744da4cc62a10886017ac694eced5a
    Key Type: RSA
    Domains: itest.run ag.itest.run test.itest.run uis.itest.run
    Expiry Date: 2022-01-27 12:17:17+00:00 (VALID: 89 days)
    Certificate Path: /etc/letsencrypt/live/itest.run/fullchain.pem
    Private Key Path: /etc/letsencrypt/live/itest.run/privkey.pem
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
```

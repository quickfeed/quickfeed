### Github Webhooks when running on localhost with self-certs

In order to test webhooks when running the server on localhost it is possible to redirect the webhook event payload to the server using [localhost.run](https://localhost.run/). It is a free service that works via ssh client and does not require any additional installations. The only downside is that given endpoits will change after a few hours and and webhook URL will need to be updated. Scm tool can be used to avoid manual updates.

Assuming the server is running on `127.0.0.1:8081`, request a tunnel from localhost.run:
```
ssh -R 80:localhost:8081 localhost.run
```

You will be given an endpoint, for example `281e3b3a0e0d02.localhost.run`. Configure webhook with the `/hook/github/events` route:

![Webhook config](./figures/webhooks-localhost.png)

To avoid editing webhook manually use scm tool:
```
make scm
scm create hook -org <organization_name> -url 'https://281e3b3a0e0d02.localhost.run/hooks/github/events' 
```

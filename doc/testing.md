# Testing gRPC-based Autograder application on itest.run VM

There are two methods to test the application on `itest.run` endpoints of `ag2` server.
Both will require you to have an active `ag2` user. If you do not have `ag2` user yet, please contact your supervisor or unix administration.
You will also need a personal endpoint on `itest.run` (with corresponding port number): several endpoints are already set up on the server, but you can add more if needed.

## Method 1: running server locally with port forwarding to itest.run server
To test application by running the server on your own machine, you will have to use ssh port forwarding. Script files `tunnel.sh` and `hometunnel.sh` are serving that purpose. `tunnel.sh` can be used when connected to **Eduroam** network, `hometunnel.sh` can be used from anywhere. 
Tunnel scrips must be adjusted before they can be used. `lport` variable is your local port where client connection from `ag2` will be forwarded. 
Replace username at the port corresponding to endpoint given to you with your `ag2` username (or add a new one).

Currently this method only works from a Chrome browser (and only tested from Debian OS).


## Method 2: running the Autograder server on itest.run VM

Make sure you start your gRPC server on a different port from `:9090` and reconfigure your instance of Envoy to listen on a port different from `:8080`, as these ports are used by the production server and proxy. Add your endpoint and webserver port to the NGINX configuration in `/etc/nginx/sites-available/default`, or use one of the existing endpoints (e.g. `test.itest.run` with port `:3006`). If adding a new endpoint, make sure to add it to the main server listener, redirecting all traffic to https:

```
server {
        listen 80;
        listen [::]:80;

        # add your new endpoint here
        server_name pedersen.itest.run meling.itest.run junaid.itest.run nicolasf.itest.run ag.itest.run test.itest.run;

        return 301 https://$host$request_uri;
}
```
Then generate a new SSl/TLS certificate with Certbot to protect the new endpoint.

gRPC client must be started with your endpoint URL (use `make remote` Makefile task to ensure recompiling with the right configuration).

Example command to start the test server on test.itest.run:
`quickfeed -service.url test.itest.run -database.file ./test.db -http.addr :3006 -http.public ./public &`



# Autograder production server

## Maintaining the server

To stop the server it is sufficient to run `killall quickfeed` command.

To update the server, run `git fetch`, then `git status` to make sure there are no local changes and the branch can be fast-forwarded.
If not, resolve local changes, preferably as new commits or pull requests to the main database. Then run `git pull`, and recompile the server (`make install`) and the client (`make ui`).

To start the server, there is a script with needed environmental variables and startup command with all necessary options provided. Run `source startag.sh` from the `quickfeed` folder to start the server.

 ## Accessing the server

 You will have to ssh to the production server `ag2.ux.uis.no` in order to be able to update or restart the server. You will need the access as an `autograder` user. After you've got the access, you will be able to ssh directly to `ag2` from the `Eduroam` network on campus. Otherwise you will have to ssh to `badne5.ux.uis.no` or `badne7.ux.uis.no` and then ssh into `ag2` from these servers.

 ## Server logs

 We use `logrotate` to maintain server logs. Configuration file is `/etc/logrotate.d/quickfeed`.
 Example configuration is:

 ```
 /home/autograder/quickfeed/ag.log {
        size 5M
        copytruncate
        dateext
        rotate 2
        compress
        maxage 14
}
```
This configuration will rotate the `ag.log` file when its size reaches 5Mb, it will start logging to a new file, will name a rotated file with the current date, will keep two latest log files compressed and will delete them after 2 week after the rotation.


[Useful logrotate manual](https://www.digitalocean.com/community/tutorials/how-to-manage-logfiles-with-logrotate-on-ubuntu-16-04)

## CI with Docker

Code submitted by students is being built and run inside Docker containers. After a container exits, the output is parsed and saved as a new submission database entry. Stopped containers are not being deleted automatically by Docker, they have to be removed manually. Right now we have a cron job removing all stopped containers every hour. 

If suddenly out of space on the production server, there are few Docker-related steps to be taken:

- check if there are containers running for too long with `docker ps`, if necessary, kill them with `docker rm <name/id>`
- check if there are too many stopped containers waiting to be removed
        - `docker ps -a` will show all container, running and stopped
        - `docker container prune` will remove all stopped containers
- restart Docker deamon with `sudo service docker restart`
- clean up all unused Docker objects with `docker system prune` (warning: can take a few minutes)

## Cron jobs

Cron is a Linux utility to schedule running of scripts or commands automatically at a specified time.

[Mininal Cron tutorial](https://www.ostechnix.com/a-beginners-guide-to-cron-jobs/).

To add, edit or remove a cron job in the user specific cron table, run `crontab -e`. 

**Important:** Cron will send the job outpus an email to every email address provided for the user. To disable emails, discard job output by adding `>/dev/null 2>&1` at the end of the job description.


## Server metrics

Statistics about connections and requests is supplied automatically by the Envoy proxy on `localhost:9901`. It is possible to access the data directly with curl by running `curl 127.0.0.1:9901/stats` in command line. The output can be formatted by adding a `format` option, e.g. `curl 127.0.0.1:9901/stats?format=json' or `curl 127.0.0.1:9901/stats?format=prometheus` or, alternatively, `curl 127.0.0.1:9901/stats/prometheus'.
Statistics about specific gRPC methods is provided by the server on `localhost:9097`.

### Prometheus
[Documentation](https://prometheus.io/docs/introduction/overview/)
Prometheus runs on port `:9095`, it scrapes metrics from the Envoy proxy and the gRPC server every 5 seconds.
To start Prometheus with all the required options run `sudo prometheus --web.listen-address="localhost:9095" --config.file=prometheus.yml --web.external-url=http://localhost:9095/stats --web.route-prefix="/" &`, or `make prometheus'.


### Grafana
[Documentation](https://grafana.com/docs/grafana/latest/)
Grafana imports the data collected by Prometheus and offers multiple visualization options. Most importantly, it can plot data from several metrics on the same graph, and also allows using predefined queries in Prometheus' query language `PromQL`. Grafana runs on the internal port `:3000` behind the NGINX proxy. It is currently available by `junaid.itest.run/grafana`. To be able to login contact a member of the Autograder team.
Configuration file to use is `etc/grafafa/grafana.ini'. Grafana will ignore any changes to the configuration file unless restarted. To restart run `sudo service grafana-server restart`.

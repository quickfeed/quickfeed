# QuickFeed Metrics Collection

Statistics about specific gRPC methods is provided by the QuickFeed via:

```sh
% curl 127.0.0.1:9097/stats
```

## Prometheus

Prometheus [documentation](https://prometheus.io/docs/introduction/overview/).

### Installing on macOS or Linux with Homebrew

```sh
% brew install prometheus
% export ETC=/usr/local/etc                  # macOS
% export ETC=/home/linuxbrew/.linuxbrew/etc  # Linux
% vim $ETC/prometheus.yml
```

Edit the file as follows:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: "prometheus"
    static_configs:
    - targets: ["localhost:9090"]

  - job_name: "quickfeed"
    static_configs:
    - targets: ["localhost:9097"]
```

If you want to access the Prometheus query interface externally, you need to edit the arguments passed to the prometheus service.

```sh
% vim $ETC/prometheus.args
```

The default is to access the query interface via localhost.

```sh
--web.listen-address=127.0.0.1:9090
```

However, here is an example with the `uis.itest.run` server.
Note that port 9090 is blocked in the campus firewall, so the Prometheus interface cannot be accessed from outside.

```sh
--web.listen-address=uis.itest.run:9090
```

After making changes to the configuration or command line arguments, you need to restart the prometheus service:

```sh
% brew services restart prometheus
```

Then you can navigate to the Prometheus query interface at [`localhost:9090`](http://localhost:9090) or [`uis.itest.run:9090`](http://uis.itest.run:9090).
Here you can search for both prometheus and quickfeed specific metrics of interest.
Here is a list of quickfeed-specific keywords:

```yaml
quickfeed_method_response_time
quickfeed_method_accessed
quickfeed_method_responded
quickfeed_method_failed
quickfeed_login_attempts
quickfeed_clone_repositories_time
quickfeed_repository_validation_time
quickfeed_test_execution_time
quickfeed_test_execution_attempts
quickfeed_test_execution_failed
quickfeed_test_execution_succeeded
```

You can also query the current aggregate statistics directly:

```sh
% curl 127.0.0.1:9097/stats
```

### Installing on Linux

TODO: Update configuration and install instructions.

Prometheus runs on port `:9095`, and scrapes metrics from the Envoy proxy and the gRPC server every 5 seconds.
To start Prometheus with all the required options run:

```sh
% make prometheus
```

## Grafana

Grafana imports the data collected by Prometheus and offers multiple visualization options.
Most importantly, it can plot data from several metrics on the same graph, and also allows using predefined queries in Prometheus' query language `PromQL`.
Grafana runs on `localhost:3000` by default.

For additional [documentation](https://grafana.com/docs/grafana/latest/).

### Installing on macOS with Homebrew

```sh
% brew install grafana
% brew services restart grafana
```

The default user and password is `admin`.
The password must be changed before using.

### Installing on Linux

TODO: Update configuration and install instructions.

It is currently available at `uis.itest.run/grafana`.
To be able to login contact a member of the QuickFeed team.

Grafana's configuration file is in `etc/grafana/grafana.ini`.
After changing the configuration file, Grafana must be restarted:

```sh
% sudo service grafana-server restart
```

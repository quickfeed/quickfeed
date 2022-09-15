# QuickFeed Metrics Collection

Statistics about specific gRPC methods is provided by the QuickFeed via:

```sh
% curl 127.0.0.1:9097/stats
```

## Prometheus

Prometheus [documentation](https://prometheus.io/docs/introduction/overview/).

### Installing on macOS with Homebrew

```sh
% brew install prometheus
% vim /usr/local/etc/prometheus.yml
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

```sh
% brew services restart prometheus
```

Navigate to the Prometheus query interface at [`localhost:9090`](http://localhost:9090).
Here you can search for both prometheus and quickfeed specific metrics of interest.
Here is a list of quickfeed-specific keywords:

```yaml
quickfeed_response_time
quickfeed_method_accessed
quickfeed_method_responded
quickfeed_method_failed
quickfeed_login_attempts
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

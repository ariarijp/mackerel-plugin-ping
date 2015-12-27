mackerel-plugin-ping
=====================

ICMP Ping RTT custom metrics plugin for mackerel.io agent.

## Usage

```shell
$ sudo ./mackerel-plugin-ping -host 8.8.8.8
```

## Example of mackerel-agent.conf

```
[plugin.metrics.ping]
command = "/path/to/mackerel-plugin-ping -host 8.8.8.8"
```

## Author

[Takuya Arita](https://github.com/ariarijp)

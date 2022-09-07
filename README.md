# go-kafka-counter

Count kafka events from multiple topics.

This small utility reads kafka messages from multiple topics and provides you a
running count of the number of seen events. If the message is a json document,
you can also break down the counts for a topic by a specific key indicated by a
JSON path.

When you quit the program pressing Ctrl-C, it also provides some statistics of
the events seen broken down by topics.

**NOTE**: This is a Golang rewrite of <https://github.com/sandipb/kafka-counter>
which I abandoned because of my application packaging requirements.

## Usage

```console
$ ./kafka-count -h
Usage of ./kafka-count:
  -c, --config string      Path to YAML config file (default "config.yaml")
  -d, --debug              Run in debug mode
  -n, --max-count int      How many messages to process before exiting. < 0 means infinite. (default -1)
  -N, --no-progress        Do not show running progress. Only the stats at the end.
  -u, --update-every int   How many ms between display updates (default 1000)
pflag: help requested
```

## Example config

```yaml
common:
  broker: kafka.default.svc.cluster.local:9092

count:

  k8sNginxIngressLogRaw:
    name: ingress            # Use this name to refer to this topic
    path: kubernetes_cluster # break down the numbers by the string value at this jsonpath location

  k8sNginxModSecurityLogRaw:
    name: modsec
    countOnly: True # Only count the total events
```

## Example Usage

Without real-time display (`-N`):

```console
$ ./kafka-count -c config.yaml -n 1000 -N

Total events received: 1000
	                       ingress:     957 (95.70%)
	                    containerd:      36 (3.60%)
	                       systemd:       6 (0.60%)
	                     k8sevents:       1 (0.10%)

Details:
                  ingress[   957] (cluster1=510(53.3)%, cluster2=447(46.7)%)
               containerd[    36] (cluster2=22(61.1)%, cluster1=14(38.9)%)
                  systemd[     6] (cluster2=3(50.0)%, cluster1=3(50.0)%)
                k8sevents[     1] (cluster1=1(100.0)%)
```

## Installation

### From source

Needs `librdkafka` development packages for compilation.

```console
$ git clone https://github.com/sandipb/go-kafka-counter.git
Cloning into 'go-kafka-counter'...
...

$ cd go-kafka-counter/

$ make
go build ./cmd/kafka-count

$ ./kafka-count -h
```

### Cross compilation

Because of its cgo requirements, the binary has to be compiled on the target OS.

If you are ok using docker for compilation, you can get a Linux binary compiled
for centos7 by running `make builddockerbuild buildindocker`.

To change the target OS version, you can edit `Dockerfile.build`.

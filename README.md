# goulding - Canary Analysis Toolkit 

Goulding is a very basic Canary Analysis Toolkit that can be used via the CLI
or as a gRPC server. It is being built with bazel and comes with a basic example.

```shell
bazel run cmd/goulding-cli -- -canaries.file $PWD/examples/config/canaries.proto.txt --canaries.request $PWD/examples/request/request.proto.txt
```

The above will run the canary locally. The requested canary will collect data
for 60s and then make a verdict. There are currently two sources implemented.
One is to HTTP GET an endpoint and the other is to run a Prometheus query. The
sources can be executed in a singleshot mode after the canary timeout or
periodically on an interval. The sources can be configured using Go templates
and variables can be passed through the canary request.

There is a primitive judge that will compare the average of each source against
a threshold. Proper statistic will be coming in a future work.

Hooks can be executed on success/failure.. once they are implemented.

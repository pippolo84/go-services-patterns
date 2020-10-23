# HProxy

HProxy is an amazing proxy that aims to lower long tail latencies.

In this repository you can also find a sample server to try it

### **server**

You can use it to experiment a server with a variable and, sometimes, **high latency** response.

```
$ curl --request GET --url http://localhost:8080

Hi, I'm the high latency server!
```
### **hproxy**

The proxy will simply do the request to the server for you, using fan-out or hedged requests to lower the tail latency.

```
$ curl --request GET --url http://localhost:8081

Hi, I'm the high latency server!
```

### What you should do:

1. complete the implementation of the **hproxy** using fan-out or hedged requests
2. try it manually to make sure that it is working
3. use the [vegeta](https://github.com/tsenart/vegeta) tool to gather info about the improved latency and compare the results with and without the proxy

### Time to spare?

Have lunch, you deserved it!
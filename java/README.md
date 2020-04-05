Aerospike Java Client Example: Using aggAPI.lua
===============================================

This project contains source code example for aggAPI.lua using the Aerospike Java Client example.

Example | Description | Link 
--- | --- | --- 
QueryLuaAgg        | Use [aggAPI.lua](https://github.com/aerospike-community/aerospike-lua-aggregations/blob/master/aggAPI.lua) to query the database using aggregation functions (min, max, count, sum)  | [View]()

#### Build

The source code can be imported into your IDE and/or built using Maven.

    mvn package

#### Run Scripts

There are two scripts to run example code:

Script | Description | Link
------ | ----------- | --- 
run_examples_swing | Run examples with a graphical user interface. | [View screenshot](http://www.aerospike.com/docs/client/java/assets/java_example_screen.png)
run_examples | Run examples on the command line. | See usage below.

#### Usage

```bash
$ ./run_examples -u

usage: com.aerospike.examples.Main [<options>] all|(<example1> <example2> ...)
options:
-d,--debug                          Run in debug mode.
-g,--gui                            Invoke GUI to selectively run tests.
-h,--host <arg>                     List of seed hosts in format:
                                    hostname1[:tlsname][:port1],...
                                    The tlsname is only used when connecting with a secure TLS
                                    enabled server. If the port is not specified, the default port
                                    is used.
                                    IPv6 addresses must be enclosed in square brackets.
                                    Default: localhost
                                    Examples:
                                    host1
                                    host1:3000,host2:3000
                                    192.168.1.10:cert1:3000,[2001::1111]:cert2:3000
-n,--namespace <arg>                Namespace (default: test)
-netty                              Use Netty NIO event loops for async examples
-nettyEpoll                         Use Netty epoll event loops for async examples (Linux only)
-P,--password <arg>                 Password
-p,--port <arg>                     Server default port (default: 3000)
-s,--set <arg>                      Set name. Use 'empty' for empty set (default: demoset)
-te,--tlsEncryptOnly                Enable TLS encryption and disable TLS certificate validation
-tls,--tlsEnable                    Use TLS/SSL sockets
-tlsCiphers,--tlsCipherSuite <arg>  Allow TLS cipher suites
                                    Values:  cipher names defined by JVM separated by comma
                                    Default: null (default cipher list provided by JVM)
-tp,--tlsProtocols <arg>            Allow TLS protocols
                                    Values:  TLSv1,TLSv1.1,TLSv1.2 separated by comma
                                    Default: TLSv1.2
-tr,--tlsRevoke <arg>               Revoke certificates identified by their serial number
                                    Values:  serial numbers separated by comma
                                    Default: null (Do not revoke certificates)
-U,--user <arg>                     User name
-u,--usage                          Print usage.
```

#### Usage Examples

    ./run_examples -h localhost -p 3000 -n test -s demoset all
    ./run_examples -h localhost -p 3000 -n test -s demoset QueryLuaAgg
    ./run_examples -g -h localhost -p 3000 -n test -s demoset

#### TLS Example

    java -Djavax.net.ssl.trustStore=TrustStorePath -Djavax.net.ssl.trustStorePassword=TrustStorePassword -jar target/aerospike-examples-*-jar-with-dependencies.jar -h "hostname:tlsname:tlsport" -tlsEnable QueryLuaAgg

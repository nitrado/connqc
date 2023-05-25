# Signal

Signal is a tool to test the network transport between multiple regions.
It consists of a server that listens for TCP probing messages, and a client that sends said messages.
The client logs the probing duration and warns if some messages get lost.

## Usage

### Server

Start the server. The default address is `:8123`:

```shell
$ signal server 
```

or specify a port:

```shell
$ signal server --addr=":1234"
```

#### More Options

The `server` command supports the following additional arguments.

```shell
OPTIONS:
   --addr value           The address to listen on for signals (default: ":8123") [$ADDR]
   --buffer-size value    The size of the read buffer used by the server (default: 512) [$BUFFER_SIZE]
   --read-timeout value   The duration after which the server should timeout when reading from a connection (default: 2s) [$READ_TIMEOUT]
   --write-timeout value  The duration after which the server should timeout when writing to a connection (default: 5s) [$WRITE_TIMEOUT]
   --log.format value     Specify the format of logs. Supported formats: 'logfmt', 'json', 'console' [$LOG_FORMAT]
   --log.level value      Specify the log level. e.g. 'debug', 'info', 'error'. (default: "info") [$LOG_LEVEL]
   --log.ctx value        A list of context field appended to every log. Format: key=value. [$LOG_CTX]
   --help, -h             show help (default: false)
```

#### Firewall

For use on firewalls managed with [`firewall-cmd`](https://firewalld.org/documentation/man-pages/firewall-cmd.html):

```shell
$ firewall-cmd --add-port=8123/tcp
# and later after the test is done
$ firewall-cmd --remove-port=8123/tcp
```

### Client

Start the client with the signal server's address:

```shell
$ signal client --addr="127.0.0.1:8123"
```

To enable more human-readable logs, set the log format:

```shell
$ signal client --addr="127.0.0.1:8123" --log.format="console"
```

To increase the amount of sent requests, reduce the interval:

```shell
$ signal client --addr="127.0.0.1:8123" --interval="200ms"
```

#### More Options

The `client` command supports the following additional arguments.

```shell
OPTIONS:
   --addr value           The address of a signal server [$ADDR]
   --backoff value        Duration to wait for before retrying to connect to the server (default: 1s) [$BACKOFF]
   --interval value       Interval at which to send probe messages to the server (default: 1s) [$INTERVAL]
   --read-timeout value   The duration after which the client should timeout when reading from a connection (default: 2s) [$READ_TIMEOUT]
   --write-timeout value  The duration after which the client should timeout when writing to a connection (default: 5s) [$WRITE_TIMEOUT]
   --log.format value     Specify the format of logs. Supported formats: 'logfmt', 'json', 'console' [$LOG_FORMAT]
   --log.level value      Specify the log level. e.g. 'debug', 'info', 'error'. (default: "info") [$LOG_LEVEL]
   --log.ctx value        A list of context field appended to every log. Format: key=value. [$LOG_CTX]
   --help, -h             show help (default: false)
```

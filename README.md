
#psdock

`psdock` is a "daemon-less" process supervisor and monitoring tool for linux. It uses [docker/libcontainer](https://github.com/docker/libcontainer) to isolate processes inside linux container.

##usage

psdock -i <image> -r <rootfs> [OPTIONS] command

#### -image, -i (required)

Specify the linux container image path in which the process run. The image is immutable, the process don't affect it in any way, it uses a "copy" of it

#### -rootfs, -r (required)

The path where the root file system of the container is created. The rootfs is a fresh copy of the image. Copies are done using the overlay union file system (mainstream since kernel 3.18). Other ways of copying the image into a rootfs can be implemented (aufs, ...)

#### -env, -e

The environment to be used by the process. This flag can be specified multiple times

#### -cwd

Current working directory of the process

#### -hostname

Container hostname

#### -user

User to use inside the container

#### -bind-mount

Host file/directory to bind mount inside the container. Format: `-bind-mount /host/path:/container/path[:ro|rw]`. This flag can be specified multiple times

#### -stdio

Setting the standard input (stdin) and outputs for the process (stdout, stderr). stdio can be interactive or not, if interactive, a tty will be available. Possible values may be:

* not specified: `psdock` uses current stdin, stdout and stderr. If it detects running inside a terminal, it allocates a tty
* `file:///path/to/output.log`: non interactive, all output are written to the provided file
* `tcp://some.tcp.server:9090`: interactive


#### -stdout-prefix

Add a prefix to each process output line. Format: `--stdout-prefix some_prefix[:color]` where color may be white, green, blue, magenta, yellow, cyan, red

#### -web-hook

Hook to call when the process state changes. Performs a HTTP PUT with this payload:

````json
{
  "ps": {
    "status": "some_status"
  }
}
````

where some_status can be: "starting", "running" or "crashed" (when the process is no longer running)

#### -bind-port

Dependent option: `-web-hook`

If the process is expected to bind a port, `psdock` will send to the web-hook the "running" status when the specified port is bound by the process or one of its children

#### -log-rotate

Dependent option: `-stdio file://*`

If given `-stdio` is a file, specifying `-log-rotate X` perform a log rotation every X hours that:

- archive (gzip) the current log file by prepending a timestamp
- empty the current log file
- keep at most 5 log archives

##Dependencies

- overlay (kernel 3.18 or over) (might work on older kernel with overlayfs, need to test)
- `-bind-port` requires `pgrep` and `lsof` to be installed on the host
- `cgroup-lites`

##psdock-ls

`psdock-ls` is a helper executable that can be used along with `psdock` (inspired by `lxc-ls`). It lists running psdock containers and display useful information:

````bash
#	CONTAINER_ID	  PID	  INIT_PID		ROOTFS		    COMMAND
0	psdock_4a59741	8988	8993		    /tmp/rootfs2	tail -f /etc/resolv.conf
1	psdock_7edd85f	8964	8969	      /tmp/rootfs1	nc -l 9999
````

* PID is `psdock` pid
* INIT_PID is the pid of the process ran by `psdock` (as seen by the host)

##Tips

- all running `psdock` containers info will be in `/var/run/psdock/*`. `psdock-ls` is here to help
- `rootfs` are ephemerals, when the process stop, they are destroyed
- `images` are immutable, there is no elegant way to create a new one so far. To do it you can spawn bash into `psdock` with the image you want to modify, make changes, and then `cp -r`  the rootfs directory before exiting from bash.
- to get images you can use [krgo](https://github.com/robinmonjo/krgo) that will give you access to images on the dockerhub (or patiently wait for [this](https://github.com/docker/distribution/tree/master/cmd/dist) to be ready)

##How signals are handled

`psdock` default behavior is to forward every signals it receives to its child process. However there are some tricky parts:

Processes are run in a container, and will have the PID 1 inside the container. Linux kernel treats PID 1 specially and will block or ignore mosts of the signals (see http://lwn.net/Articles/532748/). This means that, unless the application installs a signal handler, SIGINT and SIGTERM won't be received by the process.

`psdock` overcome this issue by inspecting the signal masks of the process to detect which signals it caught. When `psdock` received a SIGINT or a SIGTERM, it firsts check if its child process will catch it. If it does, it just forward it, otherwise, it will translate the signal to a SIGKILL.

Another solution would be to use the [phusion/baseimage-docker](https://github.com/phusion/baseimage-docker), that launch a proper init system in the container.

##Remote stdio client

`psdock` tries to allocate a tty inside the container if the `-stdio` flag is interactive. `-stdio` is considered interactive if:

1. you don't specify any stdio and launch `psdock` from a pseudo-tty (terminal, ssh ...)
2. you specify a "remote location" (for example tcp://localhost:9999 or tls://localhost:422)

It won't be considered interactive if it's launched from another program or if it points to a standard file.

In the first case, we just put the current terminal in raw mode and everything get passed to the tty inside the container (basically this means Ctrl+C and everything else can be used as is).

In the second case, obviously `psdock` do not have access to the remote terminal, so `psdock` will just disable echo on its side but won't handle Ctrl+C or anything else. A client program can be used to overcome this issue (TODO: create a sample ruby client that wait for an interactive `psdock` process to connect to it and handle everything properly)


##Roadmap

####Short term

- force kill delay (add an option to force kill the child process after some times if it didn't responds to a sigterm or sigint: killing psdock will make it's child process an orphan, we want to avoid this)
- try it with upstart
- try with overlayfs
- overlayfs hide the "work" directory
- driver for aufs and driver picking

####Medium term

- move to getgb.io ?
- handle limitation (memory / swap /cpu)
- handle OOM notification with different strategies (restart / notify ...)
- logging to syslog
- remove pgrep and lsof dependencies (http://unix.stackexchange.com/questions/131101/without-using-network-command-lines-in-linux-how-to-know-list-of-open-ports-and)
- thoughts on network namespace (libnetwork looks very promising)
- possibility to enter a running container ?

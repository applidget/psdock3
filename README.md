

#TODO now
- tests to make sure 2 processes can be ran at the same time
- implements all missing flags
- write scrip equivalent to lxc-ls with cool infos (i.e: command / pid): psdock-ls (NEW BINARY)
-

##Notion

##usage

psdock -i <image> -r <rootfs> [OPTIONS] command

-image            # required (ok)
-rootfs           # required (ok)
-stdio            # default stdin and stdout, can be file:// tcp:// tls:// ssl:// etc ... (ok)
-bind-port        # port expected to be bound (ok)
-user             # user of the process (ok)
-cwd              # cwd of the process (ok)
-web-hook         # url of the hook (ok)
-stdout-prefix    # someprefixes:green (ok)
-env              # (see how multiple args work with cli) (ok)
-hostname         # hostname (ok)
-bind-mount       # (see how multiple args work with cli) (ok)
-log-rotate       # int in hours (1) (TODO)

##dependencies

- overlay (kernel 3.18 or over)
- if -bind-port used: pgrep and lsof
- cgroup-lites

##good to know

- all running container info will be in /var/run/psdock/* (maybe create a tool to list them ?)
- rootfs are ephemeral, when the process stop, they will be destroyed
- images are immutable (to create one, spawn bash into psdock, make changes and then copy rootfs before it's destroyed)

##images

- waiting for: https://github.com/docker/distribution/tree/master/cmd/dist to be ready

##TODO

- finalize API
- proper way of options handling and checkin
- try it with upstart

##roadmap

- handle limitation (memory / swap /cpu)
- handle OOM notification with different strategies (restart / ...)
- remove pgrep and lsof dependencies (http://unix.stackexchange.com/questions/131101/without-using-network-command-lines-in-linux-how-to-know-list-of-open-ports-and)
- thoughts on network namespace (libnetwork looks very promising)
- possibility to enter a running container ?


##remote stdio client

- stty raw -echo

##how signals are handled

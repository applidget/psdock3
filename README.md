

#TODO now
- integration tests to make sure 2 processes can be ran at the same time
- integration tests log rotate ?

##Notion

##usage

psdock -i <image> -r <rootfs> [OPTIONS] command

-image, i         # required (ok)
-rootfs, r        # required (ok)
-stdio            # default stdin and stdout, can be file:// tcp:// tls:// ssl:// etc ... (ok)
-bind-port        # port expected to be bound (ok)
-user             # user of the process (ok)
-cwd              # cwd of the process (ok)
-web-hook         # url of the hook (ok)
-stdout-prefix    # someprefixes:green (ok)
-env, e           # one per env, -e available (ok)
-hostname         # hostname (ok)
-bind-mount       # multiple args possible (ok)
-log-rotate       # int in hours (1) (TODO)

##dependencies

- overlay (kernel 3.18 or over) (might work on older kernel with overlayfs, need to test)
- if -bind-port used: pgrep and lsof
- cgroup-lites

##good to know

- all running container info will be in /var/run/psdock/* (psdock-ls)
- rootfs are ephemeral, when the process stop, they will be destroyed
- images are immutable (to create one, spawn bash into psdock, make changes and then copy rootfs before it's destroyed)

##images

- waiting for: https://github.com/docker/distribution/tree/master/cmd/dist to be ready

##TODO

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

- init process in container so kernel will ign and blk some signals
- interactive process are killed

##psdock-ls

#psdock

##Notion

##API:

psdock -i /path/to/image -r /path/to/rootfs command_to_run

psdock -i /path/to/image -r /path/to/rootfs -stdio <stdio_value> command_to_run

where <stdio_value> may be:
* file:///path/to/logfile.log
* tcp://tcp.remote.com
* tls://tcp.remote.com:422
* ...

psdock -i /path/to/image -r /path/to/rootfs -prefix some_prefix:color command_to_run

where color can be
...


psdock -i image SOME_UID -h hostname command_to_run

-i immutable image
-uid if it's already running, enter the container
-h container hostname

ROADMAP

- tests, tests, tests
- stdout prefix and prefix color --> done
- log rotation --> done
- set process user --> done
- notifier --> done
- bind port option --> done
- mount rootfs --> done
- check to get dns working inside the container --> done
- cpu share / memory limit
- restart on OOM notification


- package app that conforms to heroku built app (copy in /app, source what needed and launch the app)
- possibility to enter a given container ?
- libnetwork looks pretty good

--> release and deploy and hourray !

TODO:
- tester comment ça marche avec upstart et system d (should be no problem)


A voir

- on peut avoir tout les process dans le conteneur, utiliser ça pour le port watcher ?
- si on utilise le network de l'host ça va pour le port watcher mais sinon ...

Requirements

if using bindport: pgrep and lsof

Info on remote stdin

stty raw -echo

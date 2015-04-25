#psdock

API:

psdock -i image SOME_UID -h hostname command_to_run

-i immutable image
-uid if it's already running, enter the container
-h container hostname

ROADMAP

- tests, tests, tests
- stdout prefix and prefix color --> done
- log rotation --> done
- set process user --> done
- notifier --> (done) a testé integration test ?
- bind port option --> done
- mount rootfs
- check to get dns working inside the container
- package app that conforms to heroku built app (copy in /app, source what needed and launch the app)
- possibility to enter a given container ?
- libnetwork looks pretty good

--> release and deploy and hourray !

Requirements

if using bindport: pgrep and lsof

Info on remote stdin

stty raw -echo

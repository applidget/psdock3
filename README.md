#psdock

API:

psdock -i image SOME_UID -h hostname command_to_run

-i immutable image
-uid if it's already running, enter the container
-h container hostname

ROADMAP

- tests, tests, tests
- stdout prefix and prefix color --> done
- log rotation
- check to get dns working inside the container
- set process user
- notifier co-process
- bind port option
- mount rootfs
- package app that conforms to heroku built app (copy in /app, source what needed and launch the app)
- possibility to enter a given container ?

--> release and deploy and hourray !

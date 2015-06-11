#!/bin/bash

#bash script used to test port watcher

function launch_nc {
  # the pid that binds $PORT should be the grand son of this script pid
  echo "spawning"
  nc -l $PORT
}

function prepare_nc {
  #spawn an other child process
  launch_nc &
  wait
}

#spawn a child process
prepare_nc &
wait

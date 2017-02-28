#! /bin/bash

trap 'docker-compose down' EXIT

function waitFor () {
  timeout=$(( $(date +"%s") + 60 ))  # retry window of 60 seconds
  while ! curl -s "$1" | grep "$2"; do
    if [ "$(date +"%s")" -gt "$timeout" ]; then
      echo "$1 was not ready within one minute."
      exit 1  # be sure to return a non-zero exit code on failure
    fi
    sleep 2  # wait 2 seconds before retrying
  done
}

# Tail logs
function tail () {
  docker-compose logs -f client server origin
  exit
}

# Fire up origin
docker-compose up -d origin origin_lb
docker-compose scale origin=3
waitFor localhost:81/ping PONG
if [ $1 == "origin" ]; then tail; fi

# Fire up CDN
docker-compose up -d server server_lb
docker-compose scale server=3
if [ $1 == "server" ]; then tail; fi

# Fire up client
docker-compose up -d client
docker-compose scale client=3

tail

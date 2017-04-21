#! /bin/bash

trap 'docker-compose down' EXIT

set -e  # fail if any command fails

function waitFor () {
  timeout=$(( $(date +"%s") + 60))  # retry window of 60 seconds
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
  if [[ $TESTNUM ]]; then
    curl -s localhost:8083/reset
    echo "Testing $TESTNUM clients!!!"
    sleep 60
    echo "Test $TESTNUM complete!!!"
    curl -s localhost:8083/data > data/test-$TESTNUM.json
  else
    # TODO: add --no-color if not in shell
    docker-compose logs -f client server origin ui
  fi
  exit
}

# Make sure everything is built and ready
make -j

# Fire up UI server
docker-compose up -d ui
waitFor localhost:8083/ping PONG
if [[ $1 == "ui" ]]; then tail; fi

# Fire up origin
docker-compose up -d origin origin_lb
docker-compose scale origin=3
waitFor localhost:8080/ping PONG
if [[ $1 == "origin" ]]; then tail; fi

# Fire up CDN
docker-compose up -d server server_lb
docker-compose scale server=3
waitFor localhost:8081/ping PONG
if [[ $1 == "server" ]]; then tail; fi

# Fire up client
docker-compose up -d client
if [[ $TESTNUM ]]; then
  docker-compose scale client=$TESTNUM
fi

tail

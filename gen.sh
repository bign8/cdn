#! /bin/bash
# Thanks: https://blog.codeship.com/building-minimal-docker-containers-for-go-applications/

set -e # causes the whole script to fail if any commands fail
set -o pipefail # causes pipes to fail if the input command fails

# cat > Dockerfile <<- EOM
# 	FROM scratch
# 	ADD ca-certificates.crt /etc/ssl/certs/
# 	ADD main /
# 	CMD ["/main"]
# EOM
#
# if [ ! -f certz.crt ]; then
# 	echo -n "Downloading CA Certs "
# 	curl https://curl.haxx.se/ca/cacert.pem -# -o certz.crt 2>&1 | xxd | while read; do echo -n .; done
# 	echo " Done"
# fi

# find go source directories (TODO: assert their package is main)
find */* -name "*.go" -print0 | xargs -0 -n1 dirname | sort -u | while read line; do
  pushd $line > /dev/null
  if [ ! -f ./Dockerfile ]; then
    popd > /dev/null
    continue;
  fi

  echo -n "Processing '$line' "
  # cp ../certz.crt ca-certificates.crt
  # cp ../Dockerfile Dockerfile
  #
	# # EXPOSE ui port
	# if [ "$line" == "ui" ]; then
	# 	echo "EXPOSE 9999" >> Dockerfile
	# fi

	# Verify go build
	go test | sed -n '1!p'
	echo "[Verified]"

  # Generate go source
  echo -en "\tGolang "
  CGO_ENABLED=0 GOOS=linux go build -v -installsuffix cgo -o main -ldflags="-s -w" . 2>&1 | awk getline | while read; do echo -n .; done
	# docker run --rm -e CGO_ENABLED=0 -e -v "$PWD":/usr/src/main -w /usr/src/main golang go build -v -installsuffix cgo -o main -ldflags="-s -w" 2>&1 | awk getline | while read; do echo -n .; done
  echo " Done"

	# Compressing (upx failes on OSX sierra)
	# upx main
	# docker run -v $PWD:/data --rm lalyos/upx main

  # Generating docker container
  echo -en "\tDocker "
  docker build --rm -t bign8/cdn:"$line"-latest . | while read; do echo -n .; done
  echo " Done"

  # Cleaning up assets
  echo -en "\tFinish "
  # Dockerfile ca-certificates.crt
  rm -v main | while read; do echo -n .; done
  echo " Done"

  popd > /dev/null
done
CODE=$?

# # Cleanup files
# rm Dockerfile
# docker push bign8/cdn

# Get proper exit code
if [ $CODE == 0 ]; then
	echo "Complete"
else
	echo "Failure"
	exit $CODE
fi

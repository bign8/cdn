.PHONY: client origin server
all: client origin server

CLIENT_SOURCES=client/main.go
client: $(CLIENT_SOURCES) client/main

ORIGIN_SOURCE=origin/graph.go origin/image.go origin/log.go origin/main.go origin/text.go
origin: $(CLIENT_SOURCES) origin/main

SERVER_SOURCES=server/main.go
client: $(SERVER_SOURCES) server/main

%/main:
	@printf "Processing '$(subst /main,,$@)' "
	@go test ./$(subst main,,$@) | sed -n '1!p'
	@echo "[Verified]"
	@printf "\tGolang "
	@CGO_ENABLED=0 GOOS=linux go build -v -installsuffix cgo -o $@ -ldflags="-s -w" ./$(subst main,,$@) 2>&1 | awk getline | while read; do printf "."; done
	@echo " Done!"
	@printf "\tDocker "
	@docker build --rm -t bign8/cdn:"$(subst /main,,$@)"-latest ./$(subst main,,$@) | while read; do printf "."; done
	@echo " Done!"

clean:
	@rm -f client/main
	@rm -f server/main
	@rm -f origin/main

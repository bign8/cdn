# Health

This package is designed to be called early in `main` functions.
It will allow consumers to have a container based health check.

## Usage

```
// main.go
package main

import "github.com/bign8/cdn/health"

func main() {
  health.Check()
}
```

```
# Dockerfile
FROM scratch
ADD main /
CMD ["/main"]
HEALTHCHECK --interval=1s --timeout=2s CMD ["/main", "-hc", "http://localhost:8081/ping"]
EXPOSE 8081
```

## References
- https://docs.docker.com/engine/reference/builder/#healthcheck
- https://docs.docker.com/compose/compose-file/#/healthcheck

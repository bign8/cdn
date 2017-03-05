# Documentation: https://docs.docker.com/engine/reference/builder/
FROM scratch
ADD main /
HEALTHCHECK --interval=1s --timeout=2s CMD ["/main", "-hc", "http://localhost:8082/version"]
CMD ["/main"]

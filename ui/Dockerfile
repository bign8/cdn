# Documentation: https://docs.docker.com/engine/reference/builder/
FROM scratch
ADD main /
ADD src /src
CMD ["/main"]
HEALTHCHECK --interval=2s --timeout=2s CMD ["/main", "-hc", "http://localhost:8083/version"]
EXPOSE 8083

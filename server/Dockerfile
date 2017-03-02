# Documentation: https://docs.docker.com/engine/reference/builder/
FROM scratch
ADD main /
CMD ["/main"]
HEALTHCHECK --interval=1s --timeout=2s CMD ["/main", "-hc", "http://localhost:8081/version"]
EXPOSE 8081

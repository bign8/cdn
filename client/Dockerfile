FROM scratch
ADD main /
HEALTHCHECK --interval=1s --timeout=2s CMD ["/main", "-hc", "http://localhost:8082/debug/vars"]
CMD ["/main"]

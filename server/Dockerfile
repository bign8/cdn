FROM scratch
ADD main /
CMD ["/main"]
HEALTHCHECK --interval=1s --timeout=2s CMD ["/main", "-hc", "http://localhost:8081/ping"]
EXPOSE 8081

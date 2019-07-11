
FROM scratch
EXPOSE 8080
ENTRYPOINT ["/callback-gateway"]
COPY ./bin/ /

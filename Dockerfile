FROM  gcr.io/distroless/static:nonroot

COPY connqc /connqc

ENV ADDR ":8123"

EXPOSE 8123
CMD ["./connqc", "server"]

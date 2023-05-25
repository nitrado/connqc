FROM  gcr.io/distroless/static:nonroot

COPY signal /signal

ENV ADDR ":8123"

EXPOSE 8123
CMD ["./signal", "server"]
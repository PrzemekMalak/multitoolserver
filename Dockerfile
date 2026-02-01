FROM golang:1.24-bullseye AS build

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*
RUN mkdir src/multitoolserver
WORKDIR /src/multitoolserver
COPY ./src .
RUN CGO_ENABLED=0 go build -o /bin/serv

FROM gcr.io/distroless/base-debian12
COPY --from=build /bin/serv /bin/serv

EXPOSE 8080

CMD ["/bin/serv"]

FROM --platform=linux/amd64 golang:1.24-bullseye AS build

RUN mkdir src/multitoolserver
WORKDIR /src/multitoolserver
COPY ./src .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/serv

FROM --platform=linux/amd64 scratch
COPY --from=build /bin/serv /bin/serv

EXPOSE 8080

CMD ["/bin/serv"]
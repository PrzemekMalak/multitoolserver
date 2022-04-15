FROM golang:1.18.1-bullseye AS build

RUN mkdir src/multitoolserver
WORKDIR /src/multitoolserver
COPY ./src .
RUN CGO_ENABLED=0 go build -o /bin/serv

FROM scratch
COPY --from=build /bin/serv /bin/serv

EXPOSE 8080

CMD ["/bin/serv"]
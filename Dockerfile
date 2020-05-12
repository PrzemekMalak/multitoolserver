FROM golang:1.14.2-alpine AS build

COPY serve.go ./
RUN CGO_ENABLED=0 go build -o /bin/serv

FROM scratch
COPY --from=build /bin/serv /bin/serv

EXPOSE 8080

CMD ["/bin/serv"]
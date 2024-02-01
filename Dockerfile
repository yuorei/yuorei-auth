FROM golang:1.21 as build

WORKDIR /go/src/app

RUN apt-get update -y && \
    apt-get install -y libwebp-dev

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12

COPY --from=build /go/bin/app /

EXPOSE 8080

CMD ["/app"]

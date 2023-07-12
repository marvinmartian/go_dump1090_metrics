FROM golang:1.19-alpine AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
# ENV GOARCH=arm
# ENV GOARM=6

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -a -tags netgo -ldflags '-w' -o go_dump1090_exporter 

# FROM gcr.io/distroless/static
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --chown=0:0 --from=builder /app/go_dump1090_exporter /bin/

EXPOSE 3000

CMD ["/bin/go_dump1090_exporter"]
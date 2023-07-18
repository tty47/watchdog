FROM golang:1.20.3-bullseye AS builder
WORKDIR /
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download
#COPY . .
COPY svcwatchdog /go/bin/svcwatchdog
#RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/svcwatchdog ./main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /go/bin/svcwatchdog .
ENTRYPOINT ["./svcwatchdog"]

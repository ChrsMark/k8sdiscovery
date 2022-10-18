FROM golang:1.18-alpine
WORKDIR /go/src/app

COPY main.go .
COPY  go.mod .
COPY go.sum .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o k8sdiscovery


FROM alpine:latest
COPY --from=0 /go/src/app/k8sdiscovery .
ENV PORT 80
CMD ["./k8sdiscovery"]


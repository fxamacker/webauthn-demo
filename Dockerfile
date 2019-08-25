FROM golang:1.11-alpine as builder
RUN apk add git build-base
WORKDIR /tmp/github.com/fxamacker/webauthn-demo
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o webauthn-demo

FROM alpine
WORKDIR /opt/webauthn
COPY --from=builder /tmp/github.com/fxamacker/webauthn-demo/webauthn-demo .
COPY ./static static/
EXPOSE 8443
ENTRYPOINT ["./webauthn-demo", "-addr=0.0.0.0:8443"]
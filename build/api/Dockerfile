FROM golang:1.27-alpine as builder

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o startup cmd/api/*

FROM alpine:3.24

RUN addgroup -g 1000 -S rodritv
RUN addUser -u 1000 -h /home/rodritv -S rodritv -G rodritv

USER rodritv
WORKDIR /home/rodritv

RUN mkdir config
RUN mkdir bin

COPY --chown=rodritv:rodritv --from=builder "/builder/build/startup" bin/startup
COPY --chown=rodritv:rodritv --from=builder "/builder/build/entrypoint.sh" entrypoint.sh

EXPOSE 9290
ENTRYPOINT ["sh /home/rodritv/entrypoint.sh"]
FROM golang:1.17-alpine AS build

# Install tools required
RUN apk add --no-cache git

# Copy the entire ParsecClient and build it
COPY . /go/src/ParsecClient/
WORKDIR /go/src/ParsecClient/
# use proxy to get thirdpart package in china
RUN GO111MODULE=on GOPROXY=https://goproxy.cn go get -u
# https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host
# GOARCH=arm64 or amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /bin/ParsecClient

# This results in a single layer image
FROM scratch
COPY --from=build /bin/ParsecClient /bin/ParsecClient
COPY ./ParsecClient.toml /etc/ParsecClient.toml
CMD ["/bin/ParsecClient"]

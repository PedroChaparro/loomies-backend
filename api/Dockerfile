
# -- Build --
FROM golang:1.20.10-alpine3.17 AS build
WORKDIR /source

# Install dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN go build -o /source/bin/artifact

# Use upx to compress binary
RUN apk --no-cache add git upx
RUN upx /source/bin/artifact

# -- Run --
FROM alpine:3.18.4

# Create user and set workdir
RUN adduser -D -h /opt/loomies -s /sbin/nologin -G nobody loomies
WORKDIR /opt/loomies
USER loomies

# Copy binary and run
EXPOSE 8080
COPY --from=build /source/bin/artifact /source/bin/artifact
CMD /source/bin/artifact

# -- Build --
FROM golang:1.19-alpine AS build
WORKDIR /source

# Install dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN go build -o /source/bin/artifact

# -- Run --
FROM alpine:3.17
COPY --from=build /source/bin/artifact /source/bin/artifact
CMD /source/bin/artifact
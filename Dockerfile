FROM golang:alpine AS builder

# Set necessary environmet variables needed for our image.
ENV GO111MODULE=auto \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build.
WORKDIR /build

# Gather dependencies.
RUN apk add git
RUN go get github.com/stretchr/testify
RUN go get github.com/ungerik/go-rss
RUN go get github.com/gin-gonic/gin
RUN go get github.com/google/uuid

# Copy the code into the container.
COPY . .

# Build the application.
RUN go build -o main .

# Move to /dist directory as the place for resulting binary folder.
WORKDIR /dist

# Copy binary from build to main folder.
RUN cp /build/main .

# Build a small image.
FROM scratch

# Set gin dependency to release mode, whereby logs are cleaner.
ENV GIN_MODE=release

COPY --from=builder /dist/main /

# Command to run.
ENTRYPOINT ["/main"]
# Start from golang base image
FROM golang:1.24.1-alpine3.20 AS build-stage

# Install git and build tools
RUN apk update && apk add bash && apk add build-base

WORKDIR /app

# Copy go.mod and go.sum files & Download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN --mount=type=cache,target="/go/pkg/mod" \
    CGO_ENABLED=0 go build -o /build/user ./cmd/user/main.go
# RUN CGO_ENABLED=0 go build -o /build/user ./cmd/user/main.go

# Use distroless as minimal base image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

# Copy the binary from build stage
COPY --from=build-stage /build/user /build/user

EXPOSE 8080 50051

CMD ["/build/user"] 
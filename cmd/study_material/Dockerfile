# Start from golang base image
FROM golang:1.24.1-alpine3.20 AS build-stage

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add bash && apk add build-base

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN --mount=type=cache,target="/go/pkg/mod" \
    CGO_ENABLED=0 go build -o /build/study_material ./cmd/study_material/main.go

FROM gcr.io/distroless/base-debian11 AS build-release-stage

COPY --from=build-stage /build/study_material /build/study_material

# Run the executable
CMD ["/build/study_material"]

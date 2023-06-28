FROM golang:1.20-alpine AS build
RUN apk --no-cache add git=2.40.1-r0
WORKDIR /src/
COPY go.mod ./
COPY go.sum ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /bin/service ./cmd/housekeeping/*

FROM scratch
COPY --from=build /bin/service /bin/service
USER nonroot:nonroot
ENTRYPOINT ["/bin/service"]

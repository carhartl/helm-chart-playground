FROM golang:1.20-alpine AS build
RUN apk --update add git
WORKDIR /src/
ADD . /src
RUN CGO_ENABLED=0 go build -o /bin/service

FROM scratch
COPY --from=build /bin/service /bin/service
USER nonroot:nonroot
ENTRYPOINT ["/bin/service"]

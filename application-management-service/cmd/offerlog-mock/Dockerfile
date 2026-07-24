FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /offerlog-mock ./cmd/offerlog-mock

FROM alpine:3.20
COPY --from=build /offerlog-mock /offerlog-mock
EXPOSE 9090
ENTRYPOINT ["/offerlog-mock"]

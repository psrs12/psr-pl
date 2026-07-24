FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.20
COPY --from=build /server /server
EXPOSE 8086
ENTRYPOINT ["/server"]

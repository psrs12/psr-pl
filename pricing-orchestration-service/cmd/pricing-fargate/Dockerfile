FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /pricing-fargate ./cmd/pricing-fargate

FROM alpine:3.20
COPY --from=build /pricing-fargate /pricing-fargate
ENTRYPOINT ["/pricing-fargate"]

FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /service ./cmd/contract

FROM scratch
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /service /service
COPY api/contract.yaml /app/api/contract.yaml
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/service"]

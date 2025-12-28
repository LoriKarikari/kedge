FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o kedge ./cmd/kedge

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/kedge /usr/local/bin/
ENTRYPOINT ["kedge"]
CMD ["serve"]

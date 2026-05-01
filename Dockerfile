FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o nforge .

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/nforge /nforge

EXPOSE 8080

ENTRYPOINT ["/nforge", "serve"]

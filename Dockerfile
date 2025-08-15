FROM golang:1.24 AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN go build -o server .

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/server /server
CMD ["/server"]
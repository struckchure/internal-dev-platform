FROM golang:1.25.1-alpine AS builder

RUN apk add --no-cache git

WORKDIR /code

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run github.com/steebchen/prisma-client-go generate
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /code/main .

EXPOSE 3000 9090

CMD ["./main"]

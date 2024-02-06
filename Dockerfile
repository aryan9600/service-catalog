FROM golang:alpine as builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download 

COPY cmd/ cmd/
COPY internal/ internal/
COPY docs/ docs/

RUN CGO_ENABLED=0 GOOS=linux go build -a -o main cmd/app/*

EXPOSE 8080

CMD ["./main"]

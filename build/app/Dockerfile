FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o user-service-app ./cmd/main.go

EXPOSE 8080
EXPOSE 50051

CMD ["./user-service-app"]
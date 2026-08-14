FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /qr-api .

FROM alpine:3.19
WORKDIR /app
COPY --from=build /qr-api .

EXPOSE 3000
ENTRYPOINT ["./qr-api"]

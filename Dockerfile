FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN apk add --no-cache git
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /wishlist ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /wishlist /usr/local/bin/wishlist
COPY --from=build /app/migrations ./migrations
EXPOSE 8080
CMD ["/usr/local/bin/wishlist"]

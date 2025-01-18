FROM golang:alpine AS build

WORKDIR /app



COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /output


FROM alpine:latest

WORKDIR /app


COPY --from=build /output .

EXPOSE 8080

CMD ["/app/output"]

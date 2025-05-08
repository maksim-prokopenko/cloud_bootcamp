FROM golang:1.24-alpine AS build
WORKDIR /app

#COPY go.mod go.sum ./
#RUN go mod download

COPY . .
RUN go build -o ponger-server ./internal/app/testhelpers/dummyservice/ponger

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/ponger-server .
EXPOSE 8080
CMD ["./ponger-server"]



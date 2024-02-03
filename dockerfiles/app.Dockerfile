FROM golang:1.21.6

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY main.go .

RUN go mod download

RUN go build -o bin .

ENTRYPOINT [ "/app/bin" ]
FROM golang:1.19

WORKDIR /car-service

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY go.sum .

COPY . .

RUN go build -v -o ./cmd/api ./...

CMD ["go", "run","./cmd/api"]

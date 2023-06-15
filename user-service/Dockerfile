FROM golang:1.19

WORKDIR /user-service

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o ./cmd/api ./...
EXPOSE 4001

CMD ["go","run","./cmd/api"]
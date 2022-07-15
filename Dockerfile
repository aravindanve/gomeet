FROM golang:1.18-alpine as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download && go mod verify

COPY src/ src/
RUN go build -o main ./src

# ---
FROM alpine

COPY --from=builder /workspace/main /main

ENTRYPOINT ["/main"]

FROM golang:1.26
WORKDIR /workspace
COPY . .
RUN go build ./...
CMD ["go", "test", "./..."]

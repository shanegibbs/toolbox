FROM golang:1.14 as builder

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN find .
RUN go mod download

COPY . .

ENV GOBIN /output
RUN mkdir /output
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /output ./...
RUN mv /output/sham /output/sham.linux
RUN GOOS=darwin GOARCH=amd64 go build -o /output/sham.darwin cmd/sham/main.go

FROM busybox
COPY --from=builder /output /sham/

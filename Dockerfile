FROM golang:1.14 as builder

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN find .
RUN go get -d -v ./...

COPY . .

ENV GOBIN /output
RUN go install -v ./...
# RUN env GOOS=darwin GOARCH=amd64 go build -o $GOBIN/stub-mac cmd/stub/main.go

RUN ls -al /output/

FROM busybox
COPY --from=builder /output /sham/
RUN ls -al /sham

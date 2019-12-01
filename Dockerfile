FROM golang:1.13 as builder

WORKDIR /go/src/github.com/smpio/ps-report/

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags "-s -w"


FROM scratch
COPY --from=builder /go/src/github.com/smpio/ps-report/ps-report /
ENTRYPOINT ["/ps-report"]

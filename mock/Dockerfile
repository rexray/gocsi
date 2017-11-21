#
# Test:
# $ docker run -d --rm -e CSI_ENDPOINT=tcp://:34555 -p 34555:34555 <this image>
# $ csc controller list-volumes -v 0.1.0 --endpoint tcp://127.0.0.1:3455
#
FROM golang:latest
LABEL maintainers="CSI Authors"
LABEL description="CSI Mock plugin"

WORKDIR /go
RUN go get github.com/thecodeteam/gocsi/mock
CMD ["mock"]

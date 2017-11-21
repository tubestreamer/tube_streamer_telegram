FROM golang:1.9.2
ENV  TELEGRAM_TOKEN=""
ADD . /go/src/github.com/tubestreamer/tube_streamer_telegram
RUN go get github.com/tubestreamer/tube_streamer_telegram
ENTRYPOINT /go/bin/tube_streamer_telegram

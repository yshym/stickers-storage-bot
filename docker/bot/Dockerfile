FROM golang:1.16.8

WORKDIR /home/bot
COPY . /home/bot/

RUN make build

CMD ["./bin/main"]

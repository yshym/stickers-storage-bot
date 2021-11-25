FROM golang:1.16.8

WORKDIR /code
COPY . /code/

RUN make build

CMD ["/code/bin/main"]

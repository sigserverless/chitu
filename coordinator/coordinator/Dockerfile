FROM golang:1.19-alpine as ship

RUN addgroup -S app && adduser -S -g app app
WORKDIR /root/

RUN mkdir -p /home/app

WORKDIR /home/app

COPY . .

RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN GO111MODULE="on" \
    go build -o main .

WORKDIR /home/app/
RUN chown app:app -R /home/app && chmod 777 /tmp

USER app

CMD ["./main"]

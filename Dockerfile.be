FROM golang:1.23 AS builder
RUN mkdir /tmp/build
COPY . /tmp/build
RUN cd /tmp/build && CGO_ENABLED=0 go build ./cmd/todoserv

FROM scratch
COPY --from=builder /tmp/build/todoserv /bin/todoserv
EXPOSE 2000
CMD ["/bin/todoserv"]


FROM alpine:latest

ADD build/linux/geddit /
ADD geddit.crt /
ADD geddit.key /

CMD ["/geddit"]
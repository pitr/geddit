FROM alpine:latest

ADD build/linux/geddit /
ADD geddit-prod.crt /geddit.crt
ADD geddit-prod.key /geddit.key

CMD ["/geddit"]
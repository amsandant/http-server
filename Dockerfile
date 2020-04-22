FROM alpine
MAINTAINER sandant
ADD http-server /usr/local/sandant/http-server
WORKDIR /usr/local/sandant/
ENTRYPOINT ["./http-server"]
EXPOSE 18080
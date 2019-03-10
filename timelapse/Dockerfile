FROM golang:latest
RUN apt-get update
RUN apt-get install imagemagick -y
RUN apt-get install software-properties-common -y
RUN add-apt-repository ppa:dhor/myway
RUN apt-get install dcraw -y
RUN apt-get install libimage-exiftool-perl -y
RUN apt-get install ffmpeg -y
RUN apt-get install unzip -y
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go get github.com/ddliu/go-httpclient
RUN go get github.com/satori/go.uuid
RUN go build -o main .
CMD ["/app/main"]
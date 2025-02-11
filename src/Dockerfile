FROM golang:1.17
LABEL authors="str20tbl"
RUN apt-get update
RUN apt-get install --no-install-recommends --no-install-suggests -y\
    ffmpeg sox autoconf autotools-dev automake libtool python3-pip python3-venv zip
RUN python3 -m venv /usr/.venv
RUN /usr/.venv/bin/pip3 install ffmpeg-normalize
WORKDIR /usr/src/app
RUN go install github.com/revel/cmd/revel@latest
RUN go get -t github.com/revel/revel
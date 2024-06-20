FROM golang:1.22-alpine3.18
ENV PYTHONUNBUFFERED=1
RUN apk add --no-cache make gcc musl-dev linux-headers git libstdc++-dev python3 && ln -sf python3 /usr/bin/python
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools

WORKDIR /opt

COPY . .

CMD python3 waiter.py


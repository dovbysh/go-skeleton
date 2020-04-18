FROM golang:1.14.2-buster
WORKDIR /opt
RUN go get -v github.com/rubenv/sql-migrate/...
COPY dbconfig.yml .
COPY migrations migrations

CMD sql-migrate up -env=db

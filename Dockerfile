FROM golang:1.14.2-buster as build
WORKDIR /opt
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM golang:1.14.2-buster as prod
WORKDIR /opt

COPY --from=build /opt/bin/skeleton /opt/bin/skeleton
COPY --from=build /opt/configs/config_docker.yaml /opt/configs/config.yaml
COPY --from=build /opt/api /opt/api

EXPOSE 80
CMD ["/opt/bin/skeleton", "-c", "./configs/config.yaml", "-swagger", "./api/openapi_spec"]


FROM golang:1.21.5-alpine3.19 AS build

WORKDIR /api

COPY ./ .
RUN go build -trimpath -v -o build/api ./cmd/main.go

FROM golang:1.21.5-alpine3.19 AS api

WORKDIR /api

COPY --from=build /api/build .

CMD ["./api"]

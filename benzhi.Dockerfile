FROM golang:1.23.12 AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o /fuelhydrant ./cmd/fuelhydrant

FROM golang:1.23.12
ENV GOPROXY=off GOSUMDB=off
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
COPY --from=build /fuelhydrant /usr/local/bin/fuelhydrant
EXPOSE 8091
CMD ["/usr/local/bin/fuelhydrant", "-addr", "0.0.0.0:8091", "-dir", "/app/data"]

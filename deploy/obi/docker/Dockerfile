FROM golang:1.27.3 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /obi-example ./cmd/obi-example

FROM gcr.io/distroless/static-debian12
COPY --from=build /obi-example /obi-example
ENTRYPOINT ["/obi-example"]

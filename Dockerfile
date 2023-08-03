FROM --platform=linux/amd64 golang:1.20 as build
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /go/app
COPY go.* ./
RUN go mod download
COPY . .
RUN make build

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/app/bin /app
ENTRYPOINT ["/app/dispatcher"]

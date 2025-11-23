FROM golang:latest AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make generate && (CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /src/pr_splitter ./cmd/pr_splitter/)

FROM alpine:latest AS app
WORKDIR /application
COPY --from=build /src/pr_splitter .
CMD ["./pr_splitter"]
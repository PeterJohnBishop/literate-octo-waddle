FROM golang:1.24-alpine

WORKDIR /app
RUN adduser -D -g '' appuser

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# build the binary with a unique name
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /app/app-binary .

EXPOSE 8081
USER appuser

CMD ["/app/app-binary"]
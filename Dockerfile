FROM public.ecr.aws/docker/library/golang:1.21.7-bookworm as builder
WORKDIR /app

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o rinha cmd/socket/main.go

FROM public.ecr.aws/docker/library/alpine:latest
WORKDIR /usr/bin

COPY --from=builder /app/rinha .

CMD [ "rinha" ]

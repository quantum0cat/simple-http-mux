# Start from golang:1.12-alpine base image
FROM golang:1.18-alpine

# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
RUN apk update && apk upgrade && apk add --no-cache bash git openssh

LABEL maintainer="John Doe <quantum0cat@gmail.com>"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o simple_http_mux ./cmd/main/app.go
RUN ls -al .

EXPOSE 10000

CMD ["./simple_http_mux"]
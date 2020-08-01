FROM golang:1.14-alpine as build-env

WORKDIR /go/bin
COPY go.mod .
COPY go.sum .


RUN go mod download
COPY . .

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
        go build -ldflags "-s -w -X 'main.version=${revision}'" \
        -o /starish \
        cmd/starish/starish.go 


FROM scratch
COPY --from=build-env /starish /starish
CMD ["/starish"]
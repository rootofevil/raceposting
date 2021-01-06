FROM golang:1.14 AS build
WORKDIR /go/src/github.com/rootofevil/raceposting
RUN go get -d -v github.com/golang/freetype github.com/hqbobo/text2pic github.com/huandu/facebook
COPY *.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /go/src/github.com/rootofevil/raceposting/app .
CMD ["./app"]  
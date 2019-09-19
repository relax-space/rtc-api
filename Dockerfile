FROM pangpanglabs/golang:builder-beta AS builder
WORKDIR /go/src/rtc-api
COPY . .
# disable cgo
ENV CGO_ENABLED=0
# build steps
RUN echo ">>> 1: go version" && go version
RUN echo ">>> 2: go get" && go get -v -d
RUN echo ">>> 3: go install" && go install
 
# make application docker image use alpine
FROM pangpanglabs/alpine-ssl
WORKDIR /go/bin/
# copy config files to image
COPY --from=builder /go/src/rtc-api/*.yml ./
# copy execute file to image
COPY --from=builder /go/bin/rtc-api ./rtc-api
RUN chmod +x ./rtc-api

EXPOSE 8080
CMD ["./rtc-api"]
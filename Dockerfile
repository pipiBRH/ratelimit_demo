# build stage
FROM golang:1.12-stretch AS build-env
ADD . /work/
WORKDIR /work
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /work/app
 
# final stage
FROM scratch
COPY --from=build-env /work/app /work/app
WORKDIR /work
EXPOSE 4000
CMD ["/work/app"]
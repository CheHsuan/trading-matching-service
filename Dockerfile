# build cmd stage
FROM golang:1.17-alpine3.15 as build_cmd
WORKDIR /repo
COPY . .
RUN go build -o /bin/app .

# build runtime stage
FROM alpine:3.15
WORKDIR /cmd
COPY --from=build_cmd /bin/app /cmd/app
ENTRYPOINT ["/cmd/app"]

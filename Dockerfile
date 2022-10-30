FROM golang:latest AS build 
WORKDIR /code
COPY . .
ENV CGO_ENABLED 0
RUN go mod download
RUN go build -o server ./cmd/main.go
# RUN echo "$(ls -R)" && exit 1

FROM alpine:latest
WORKDIR /backend
COPY --from=build /code/server .
COPY --from=build /code/internals/storage/init.sql internals/storage/
EXPOSE 8080
ENTRYPOINT ["./server"]
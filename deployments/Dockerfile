FROM golang:1.19-alpine3.18 as build
WORKDIR /app
COPY go.mod go.sum ./
COPY ./cmd /app/cmd
COPY ./internal /app/internal
COPY ./pkg /app/pkg
COPY ./vendor /app/vendor
COPY ./fixtures /app/fixtures
RUN cd cmd/storageapi && go build -o app

FROM alpine:3.18 as prod
COPY --from=build /app/cmd/storageapi/app ./app
COPY --from=build /app/fixtures ./fixtures
EXPOSE 3001
ENTRYPOINT ["/app"]
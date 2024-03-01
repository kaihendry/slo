FROM golang:latest AS build-env

RUN mkdir -p /workspace
WORKDIR /workspace

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

COPY . .

RUN go build

FROM scratch

COPY --from=build-env /workspace/slo /

EXPOSE 8080
ENV PORT 8080

ENTRYPOINT ["/slo"]
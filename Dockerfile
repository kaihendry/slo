FROM golang as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ARG TARGET_ARCH=amd64
RUN echo Building for ${TARGET_ARCH} && CGO_ENABLED=0 GOOS=linux GOARCH=${TARGET_ARCH} go build

FROM scratch
ENV PORT 9000
COPY --from=builder /app/sla /app/
ENTRYPOINT ["/app/sla"]

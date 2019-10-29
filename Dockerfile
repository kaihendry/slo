FROM golang as builder

ENV GO111MODULE=on
ARG GOARM=""

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
ENV CGO_ENABLED=0
ARG TARGET_ARCH=amd64
RUN echo Building for ${TARGET_ARCH} ${GOARM}
RUN GOOS=linux GOARM=${GOARM} GOARCH=${TARGET_ARCH} go build -ldflags "-s -w" -a -installsuffix cgo

FROM scratch
ENV PORT 9000
COPY --from=builder /app/sla /app/
ENTRYPOINT ["/app/sla"]

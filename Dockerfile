FROM golang as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ARG TARGET_ARCH=amd64
ARG VERSION
ARG BRANCH
ARG USER
ARG BUILDDATE
ARG HOST

LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.version=$VERSION

RUN echo Building for ${TARGET_ARCH}
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGET_ARCH} \
	go build -ldflags "-X main.Version=${VERSION} \
		-X main.Branch=${BRANCH} \
		-X main.BuildDate=${BUILDDATE} \
		-X main.BuildUser=${USER}@${HOST}"

FROM scratch
ENV PORT 9000
COPY --from=builder /app/sla /app/
ENTRYPOINT ["/app/sla"]

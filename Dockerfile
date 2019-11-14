FROM golang as builder

WORKDIR /app

COPY . .

ARG TARGET_ARCH=amd64
ARG VERSION
ARG BRANCH
ARG USER
ARG HOST

LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.version=$VERSION

RUN echo Building for ${TARGET_ARCH}
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGET_ARCH} \
	go build -ldflags "-X main.Version=${VERSION} \
		-X main.Branch=${BRANCH} \
		-X main.BuildUser=${USER}@${HOST}"

FROM scratch
COPY --from=builder /app/sla /app/
ENTRYPOINT ["/app/sla"]

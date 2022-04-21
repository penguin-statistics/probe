FROM golang:1.18.1-alpine AS base
WORKDIR /app

# builder
FROM base AS gobuilder
ENV GOOS linux
ENV GOARCH amd64

# go 1.18 requires git
RUN apk add --no-cache git

# modules: utilize build cache
COPY go.mod ./
COPY go.sum ./

# RUN go env -w GO111MODULE=on && go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . .

# build the binary
RUN go build -o probe .

# runner
FROM base AS runner
RUN apk add --no-cache libc6-compat tini
# Tini is now available at /sbin/tini

COPY --from=gobuilder /app/probe /app/probe

ENTRYPOINT ["/sbin/tini", "--"]
CMD [ "/app/probe" ]

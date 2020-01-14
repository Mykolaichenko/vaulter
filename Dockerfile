FROM golang:1.13.0-alpine3.10

ENV GO111MODULE on
ENV PROJECT_NAME vaulter

RUN apk --no-cache add git

WORKDIR /go/src/${PROJECT_NAME}

COPY . .

RUN mkdir -p /app

RUN VERSION=$(git describe --always --long) && \
    DT=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
    SEMVER=$(git tag --list --sort="v:refname" | tail -n -1) && \
    BRANCH=$(git rev-parse --abbrev-ref HEAD) && \
    go mod init && go get . && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s -X main.version=${VERSION} -X main.builddt=${DT} -X main.semver=${SEMVER} -X main.branch=${BRANCH}" -o /app/${PROJECT_NAME}

ENTRYPOINT /app/vaulter


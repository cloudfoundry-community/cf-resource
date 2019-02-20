FROM golang:alpine as builder
RUN apk add --no-cache curl jq
RUN mkdir -p /assets
WORKDIR /assets
RUN curl -L "https://packages.cloudfoundry.org/stable?release=linux64-binary&source=github" | tar -xzf -
COPY . /go/src/github.com/concourse/cf-resource
ENV CGO_ENABLED 0
RUN go build -o /assets/in github.com/concourse/cf-resource/in/cmd/in
RUN go build -o /assets/out github.com/concourse/cf-resource/out/cmd/out
RUN go build -o /assets/check github.com/concourse/cf-resource/check/cmd/check
WORKDIR /go/src/github.com/concourse/cf-resource
RUN set -e; for pkg in $(go list ./... | grep -v "acceptance"); do \
		go test -o "/tests/$(basename $pkg).test" -c $pkg; \
	done

FROM alpine:edge AS resource
RUN apk add --no-cache bash tzdata ca-certificates
COPY --from=builder assets/ /opt/resource/
RUN chmod +x /opt/resource/*
RUN mv /opt/resource/cf /usr/bin/cf

FROM resource AS tests
COPY --from=builder /tests /go-tests
COPY out/assets /go-tests/assets
WORKDIR /go-tests
RUN set -e; for test in /go-tests/*.test; do \
		$test; \
	done

FROM resource

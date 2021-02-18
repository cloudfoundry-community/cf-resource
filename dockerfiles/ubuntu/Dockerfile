ARG base_image
ARG builder_image=concourse/golang-builder

FROM ${builder_image} as builder
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    jq \
  && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /assets
WORKDIR /assets
RUN curl -L "https://packages.cloudfoundry.org/stable?release=linux64-binary&source=github" | tar -xzf -
WORKDIR /src
COPY . .
RUN go mod download
RUN go build -o /assets/in github.com/concourse/cf-resource/in/cmd/in
RUN go build -o /assets/out github.com/concourse/cf-resource/out/cmd/out
RUN go build -o /assets/check github.com/concourse/cf-resource/check/cmd/check
RUN set -e; for pkg in $(go list ./... | grep -v "acceptance"); do \
		go test -o "/tests/$(basename $pkg).test" -c $pkg; \
	done

FROM ${base_image} AS resource
RUN apt-get update && apt-get install -y --no-install-recommends \
    tzdata \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*
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

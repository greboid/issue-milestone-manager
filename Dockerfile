FROM ghcr.io/greboid/dockerfiles/golang as build

# Build the app
WORKDIR /app
COPY . /app
#Compile the app. Retrieves licenses, set timestamps on the outputs
RUN set -eux; \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -gcflags=./dontoptimizeme=-N -ldflags=-s -o /go/bin/app .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    mkdir /data; \
    touch --date=@0 /go/bin/app /notices /data

FROM ghcr.io/greboid/dockerfiles/base
COPY --from=build /notices /notices
COPY --from=build /go/bin/app /issue-tagger
WORKDIR /
CMD ["/issue-tagger"]

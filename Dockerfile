FROM ghcr.io/greboid/dockerfiles/golang@sha256:65e504b0cb4e5df85e2301f47cd3f231768d7b0d5aba59b1201e9c50fdf5e0ac AS BUILD

# Build the app
WORKDIR /app
COPY . /app
#Compile the app. Retrieves licenses, set timestamps on the outputs
RUN set -eux; \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -gcflags=./dontoptimizeme=-N -ldflags=-s -o /go/bin/app .; \
    go run github.com/google/go-licenses@latest save ./... --save_path=/notices; \
    mkdir /data; \
    touch --date=@0 /go/bin/app /notices /data

FROM gcr.io/distroless/static:nonroot@sha256:c9f9b040044cc23e1088772814532d90adadfa1b86dcba17d07cb567db18dc4e
COPY --from=build /notices /notices
COPY --from=build /go/bin/app /issue-tagger
WORKDIR /
CMD ["/issue-tagger"]

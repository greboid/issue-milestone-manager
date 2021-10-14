FROM docker.io/golang:1.17.2@sha256:cf615c1499e8bc2cf00696ba234cddd47fdb8d9a3b37b7c35726e46ee4ae08cc AS build

# Build the app
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -gcflags=./dontoptimizeme=-N -ldflags=-s -o /go/bin/app .
RUN mkdir /data

# Generate licence information
RUN go get github.com/google/go-licenses && go-licenses save ./... --save_path=/notices

FROM gcr.io/distroless/static:nonroot@sha256:07869abb445859465749913267a8c7b3b02dc4236fbc896e29ae859e4b360851
COPY --from=build /notices /notices
COPY --from=build /go/bin/app /issue-tagger
WORKDIR /
CMD ["/issue-tagger"]

FROM golang

# Copy the local package files to the container's workspace.
ADD ./src/server /go/src/github.com/golang/server

RUN go install github.com/golang/server

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/server

# Document that the service listens on port 8080.
EXPOSE 8080
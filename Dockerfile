# Use the latest version from the 1.12 tree.
FROM golang:1.12

LABEL maintainer="Avishkar Gupta <avgupta@redhat.com>"

# Copy the entire source code to the image
COPY ./ /apps/

WORKDIR /apps/
# Update dependencies
RUN go mod tidy

RUN go mod vendor

# Now build the source code
RUN make build && make install


EXPOSE 8080

CMD ["/go/bin/api"]

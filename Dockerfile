FROM golang:latest-alpine as build

# Installing Git
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# Build the code
RUN go build consumer.go

# Create Docker container for application
FROM alpine as runtime

# Copy the executable to app folder
COPY --from=build $GOPATH/src/consumer $GOPATH/app/image-service

# Run the executable
CMD ["$GOPATH/app/image-service"]

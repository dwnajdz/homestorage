# Use the official Golang 1.19 image as the base image
FROM golang:1.19

# Create app folder 
RUN mkdir /app

# Set the working directory inside the container
WORKDIR /app

# Copy the source code to the working directory
COPY src/ .

# Fetch the dependencies using go mod
RUN go mod download

# Build the Go application
RUN go build -o main .

# Expose port 3000 for the HTTP server
EXPOSE 3000

# Set the entry point command for the container
CMD ["./main"]

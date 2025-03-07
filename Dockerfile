FROM golang:1.23.4-alpine
# Install required packages
RUN apk add --no-cache git build-base && \
    go install github.com/air-verse/air@latest  

# Set working directory
WORKDIR /app
# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

RUN touch build-errors.log && \
    chmod 666 build-errors.log
# Expose ports
EXPOSE 8080 50051

# Run the application
CMD ["air", "-c", ".air.docker.toml"]
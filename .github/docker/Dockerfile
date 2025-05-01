FROM alpine:3.18

# Install ca-certificates for secure connections
RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

# Create a non-root user and group
RUN addgroup -S globus && \
    adduser -S globus -G globus

# Create app directory
WORKDIR /app

# Copy the pre-built binary
COPY globus /app/globus

# Set binary as executable
RUN chmod +x /app/globus

# Use non-root user
USER globus

# Set the binary as the entrypoint
ENTRYPOINT ["/app/globus"]

# Command will be passed as arguments to the entrypoint
CMD ["--help"]
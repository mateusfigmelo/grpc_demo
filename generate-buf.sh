#!/bin/bash

# Generate code using Buf
echo "Generating protobuf files with Buf..."

# Download dependencies
buf dep update library

# Generate code
buf generate

echo "Protobuf files generated successfully!" 
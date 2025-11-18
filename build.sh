#!/bin/bash
# Build script for Multitool Server Docker image
# Default: x86_64/amd64 architecture

set -e

IMAGE_NAME="${IMAGE_NAME:-przemekmalak/multitoolserver}"
ARCH="${ARCH:-amd64}"
TAG="${TAG:-latest}"

echo "Building Docker image: ${IMAGE_NAME}:${TAG} for architecture: ${ARCH}"

case "${ARCH}" in
  amd64|x86_64)
    PLATFORM="linux/amd64"
    ;;
  arm64|aarch64)
    PLATFORM="linux/arm64"
    echo "Note: Dockerfile is configured for amd64. For ARM64, you may need to modify the Dockerfile."
    ;;
  *)
    echo "Unsupported architecture: ${ARCH}"
    echo "Supported: amd64, x86_64"
    exit 1
    ;;
esac

docker build \
  --platform "${PLATFORM}" \
  -t "${IMAGE_NAME}:${TAG}" \
  -t "${IMAGE_NAME}:${TAG}-${ARCH}" \
  .

echo "Build complete!"
echo "Image: ${IMAGE_NAME}:${TAG}"
echo "Image: ${IMAGE_NAME}:${TAG}-${ARCH}"


#!/bin/bash

# 檢查參數數量
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version_number>"
    exit 1
fi

# 構建 api image
docker buildx build \
    --platform linux/arm64 \
    -f ./.ci/api.Dockerfile \
    -t q4-api:$1 \
    .
if [ $? -ne 0 ]; then
    echo "Failed to build api image."
    exit 1
fi


# 構建 ui image
docker buildx build \
    --platform linux/arm64 \
    -f ./.ci/ui.Dockerfile \
    -t q4-ui:$1 \
    .
if [ $? -ne 0 ]; then
    echo "Failed to build ui image."
    exit 1
fi

# 打包 helm chart
helm package ./.ci/helm
if [ $? -ne 0 ]; then
    echo "Failed to package helm chart."
    exit 1
fi


# push images
docker tag q4-api:$1 ghcr.io/ljwun/q4-api:$1
docker tag q4-api:$1 ghcr.io/ljwun/q4-api:latest
docker tag q4-ui:$1 ghcr.io/ljwun/q4-ui:$1
docker tag q4-ui:$1 ghcr.io/ljwun/q4-ui:latest

docker push ghcr.io/ljwun/q4-api:$1
if [ $? -ne 0 ]; then
    echo "Failed to push q4-api:$1."
    exit 1
fi
docker push ghcr.io/ljwun/q4-api:latest
if [ $? -ne 0 ]; then
    echo "Failed to push q4-api:latest."
    exit 1
fi
docker push ghcr.io/ljwun/q4-ui:$1
if [ $? -ne 0 ]; then
    echo "Failed to push q4-ui:$1."
    exit 1
fi
docker push ghcr.io/ljwun/q4-ui:latest
if [ $? -ne 0 ]; then
    echo "Failed to push q4-ui:latest."
    exit 1
fi

# push helm chart
helm push q4-$1.tgz oci://ghcr.io/ljwun/helm
if [ $? -ne 0 ]; then
    echo "Failed to push helm chart."
    exit 1
fi

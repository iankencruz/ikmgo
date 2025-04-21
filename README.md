# ikmgo

# Build command

docker buildx build \
 --platform linux/amd64 \
 -t iankendoit/ikmgo:latest \
 -t iankendoit/ikmgo:v1.0.2 \
 --push .

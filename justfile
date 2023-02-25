bisu:
    CC=$(which musl-gcc) go build -ldflags='-s -w -linkmode external -extldflags "-static"' -o ./bisu

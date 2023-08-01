run-local:
    go build && godotenv ./bisu

verbose:
    go build -tags debug && godotenv ./bisu

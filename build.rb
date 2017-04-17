from "golang"

TARGET = "/go/src"
REPO = "github.com/box-builder/tarutil"

path = "#{TARGET}/#{REPO}"

copy ".", path
workdir path
run "go get -t -d ./..."

# this entrypoint clear works around a box 0.5.1 bug.
set_exec entrypoint: [], cmd: ["/bin/sh", "-c", "go test -cover -v ./..."]

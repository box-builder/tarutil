from "golang"

TARGET = "/go/src"
REPO = "github.com/unclejack/tarutil"

path = "#{TARGET}/#{REPO}"

copy ".", path

set_exec cmd: ["/bin/sh", "-c", "cd #{path} && go test -v ./..."]

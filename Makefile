# This how we want to name the binary output
BINARY=raylib-go-mines

# These are the values we want to pass for VERSION and BUILD
# git tag 1.0.1
# git commit -am "One more change after the tags"
VERSION=`git describe --tags | cut -d- -f1`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION}"

# Builds the project
build:
	go build ${LDFLAGS} -o ${BINARY}
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY}.exe

# Runs the project
run:
	go run ${LDFLAGS} .

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean install

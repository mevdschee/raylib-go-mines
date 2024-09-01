# This how we want to name the binary output
BINARY=raylib-go-mines

# These are the values we want to pass for VERSION and BUILD
# git tag 1.0.1
# git commit -am "One more change after the tags"
VERSION=`git describe --tags | cut -d- -f1 | sed 's/[^0-9.]*//g'`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGSANY=-ldflags "-w -s -X main.version=${VERSION}"
LDFLAGSWIN=-ldflags "-w -s -X main.version=${VERSION} -H=windowsgui"

# Builds the project
build:
	go build ${LDFLAGSANY} -o ${BINARY}
	cat winres/winres_template.json | sed s/__VERSION__/${VERSION}/g > winres/winres.json
	~/go/bin/go-winres make
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build ${LDFLAGSWIN} -o ${BINARY}.exe

# Runs the project
run:
	go run .

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean install

BUILD_DIR=./bin
VERSION=${GIT_COMMIT_HASH}
DATE=`date +%FT%T%z`

LDFLAGS=-ldflags "-X main.buildVersion=${VERSION} -X main.buildDate=${DATE}"
MAIN=main.go
OUT=step-file-viewer

ENV=env CGO_ENABLED=0
ENV_WINDOWS=${ENV} GOOS=windows
ENV_LINUX=${ENV} GOOS=linux
ENV_MACOS=${ENV} GOOS=darwin

all: 
	$(MAKE) build

clean:
	rm -rf ${BUILD_DIR}

build: windows linux macos

mac:
	go build ${LDFLAGS} -o ${BUILD_DIR}/mac/${OUT} ${MAIN}

windows:
	${ENV_WINDOWS} GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/windows/${OUT}-amd64.exe ${MAIN}

linux:
	${ENV_LINUX} GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/linux/${OUT}-amd64 ${MAIN}

macos:
	${ENV_LINUX} GOARCH=amd64 go build ${LDFLAGS} -o ${BUILD_DIR}/macos/${OUT}-amd64 ${MAIN}	

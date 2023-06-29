BINARY_NAME=monobrew
BINARY=dist/${BINARY_NAME}

build:
	go build -o ${BINARY} -ldflags="-s -w" cmd/monobrew/main.go

run: build
	${BINARY}

clean:
	go clean
	rm -f ${BINARY}

publish: ${BINARY}
	if [ "${PUBLISH_ROOT}" != "" ]; then scp ${BINARY} conf/* ${PUBLISH_ROOT}; fi

.PHONY: build run

MAIN_GO=./cmd/main.go
BUILD_PATH=build
BINARY_NAME=storage_management
BUILD_FLAGS=-ldflags "-s -w"

build:
	@CGO_ENABLED=0 go build ${BUILD_FLAGS} -o ${BUILD_PATH}/${BINARY_NAME} ${MAIN_GO}

run: build
	@./${BUILD_PATH}/${BINARY_NAME}

clean:
	@rm -f ${BUILD_PATH}/${BINARY_NAME}
	@echo "done cleaning..."

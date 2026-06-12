.PHONY: build run

MAIN_GO=./cmd/main.go
BUILD_PATH=build
BINARY_NAME=storage_management

build:
	@go build -o ${BUILD_PATH}/${BINARY_NAME} ${MAIN_GO}

run: build
	@./${BUILD_PATH}/${BINARY_NAME}

clean:
	@rm -f ${BUILD_PATH}/${BINARY_NAME}
	@echo "done cleaning..."

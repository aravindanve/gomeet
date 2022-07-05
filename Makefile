INPUT=./src
OUTPUT=./bin/server

all: build test

build:
	go build -o ${OUTPUT} ${INPUT}

clean:
	go clean
	rm -f ${OUTPUT}

test:
	go test ${INPUT}/...

run:
	make build
	./${OUTPUT}

watch:
	if [ -z `which gow` ]; then echo "installing gow..."; go install github.com/mitranim/gow@latest; echo "done!"; fi
	gow -e="go,mod,html" run ./src

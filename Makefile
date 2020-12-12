VANTA = ./main.go

all: run-all


run:
	go run -race ${VANTA}


# if source tree cloned
repl:
	rlwrap go run -race ${VANTA} -i \
		../klisp/lib/klisp.klisp ../klisp/lib/math.klisp


# if source tree cloned 
run-all:
	go run -race ${VANTA} \
		../klisp/lib/klisp.klisp ../klisp/lib/math.klisp \
		../klisp/test/*.klisp


# build for specific OS target
build-%:
	GOOS=$* GOARCH=amd64 go build -o klisp-$* ${VANTA}


build:
	go build -o klisp ${VANTA}


install:
	go build -o $$GOBIN/vanta


# clean any generated files
clean:
	rm -rvf klisp klisp-*


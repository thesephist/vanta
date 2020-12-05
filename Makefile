VANTA = ./main.go

all: run-all


run:
	go run -race ${VANTA}


# if source tree cloned
repl:
	rlwrap go run -race ${VANTA} -i ../klisp/lib/klisp.klisp


# if source tree cloned 
run-all:
	go run -race ${VANTA} ../klisp/lib/klisp.klisp ../klisp/test/*.klisp


# build for specific OS target
build-%:
	GOOS=$* GOARCH=amd64 go build -o klisp-$* ${VANTA}


build:
	go build -o klisp ${VANTA}


# clean any generated files
clean:
	rm -rvf klisp klisp-*


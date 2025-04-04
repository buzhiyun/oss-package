GOEXEC = go
CGO = CGO_ENABLED=0
GOARCH = GOARCH=amd64
LINUX_GOOS = GOOS=linux
WINDOWS_GOOS = GOOS=windows
BINARY = package-oss
LDFLAGS := -s -w


.PHONY: build start

# build-linux: 
build-linux:
	GOPROXY=https://goproxy.cn,direct GO111MODULE=on ${GOEXEC} mod vendor
	${CGO} ${GOARCH} ${LINUX_GOOS} ${GOEXEC} build  -mod=mod  -ldflags "$(LDFLAGS)" -o bin/${BINARY} cmd/package-oss/main.go

build-windows:
	GOPROXY=https://goproxy.cn,direct GO111MODULE=on ${GOEXEC} mod vendor
	${CGO} ${GOARCH} ${WINDOWS_GOOS} ${GOEXEC} build  -mod=mod  -ldflags "$(LDFLAGS)" -o bin/${BINARY}.exe cmd/package-oss/main.go

build-all: build-linux build-windows


package-linux: build-linux
	cp config.yaml bin/
	mkdir -p output
	tar -C bin/ -cJvf output/${BINARY}_linux_x86_64.tar.xz ${BINARY} config.yaml
	rm -rf bin/config.yaml

package-windows: build-windows
	cp config.yaml bin/
	mkdir -p output
	tar -C bin/ -cJvf output/${BINARY}_windows_x86_64.tar.xz ${BINARY}.exe config.yaml
	rm -rf bin/config.yaml


package: package-linux package-windows

start:
	${GOEXEC} run -mod=mod  .

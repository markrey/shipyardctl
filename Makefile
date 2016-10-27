#Format is MAJOR . MINOR . PATCH

VERSION=1.3.0
GO_VERSION=1.7

release: dir-build package-linux package-windows package-darwin

package-linux: dir-linux build-linux tar-linux

package-windows: dir-windows build-windows tar-windows

package-darwin: dir-darwin build-darwin tar-darwin

dir-build:
	mkdir build

build-linux:
	GOOS=linux GOARCH=amd64 go build -o build/linux/shipyardctl

dir-linux:
	mkdir build/linux

tar-linux:
	tar -zcvf build/linux/shipyardctl-$(VERSION).linux.amd64.go$(GO_VERSION).tar.gz build/linux/shipyardctl

build-windows:
	GOOS=windows GOARCH=amd64 go build -o build/windows/shipyardctl.exe

dir-windows:
	mkdir build/windows

tar-windows:
	tar -zcvf build/windows/shipyardctl-$(VERSION).windows.amd64.go$(GO_VERSION).tar.gz build/windows/shipyardctl.exe

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o build/darwin/shipyardctl

tar-darwin:
	tar -zcvf build/darwin/shipyardctl-$(VERSION).darwin.amd64.go$(GO_VERSION).tar.gz build/darwin/shipyardctl

dir-darwin:
	mkdir build/darwin

clean:
	rm -r build
#Format is MAJOR . MINOR . PATCH

VERSION=1.3.1
GO_VERSION=1.7

release: dir-build package-linux package-windows package-darwin

package-linux: dir-linux build-linux zip-linux

package-windows: dir-windows build-windows zip-windows

package-darwin: dir-darwin build-darwin zip-darwin

dir-build:
	mkdir build

build-linux:
	GOOS=linux GOARCH=amd64 go build -o build/linux/shipyardctl

dir-linux:
	mkdir build/linux

zip-linux:
	zip -r build/linux/shipyardctl-$(VERSION).linux.amd64.go$(GO_VERSION).zip build/linux/shipyardctl

build-windows:
	GOOS=windows GOARCH=amd64 go build -o build/windows/shipyardctl.exe

dir-windows:
	mkdir build/windows

zip-windows:
	zip -r build/windows/shipyardctl-$(VERSION).windows.amd64.go$(GO_VERSION).zip build/windows/shipyardctl.exe

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o build/darwin/shipyardctl

zip-darwin:
	zip -r build/darwin/shipyardctl-$(VERSION).darwin.amd64.go$(GO_VERSION).zip build/darwin/shipyardctl

dir-darwin:
	mkdir build/darwin

clean:
	rm -r build
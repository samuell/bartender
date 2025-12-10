build: bartender

bartender:
	go build

build-static: fyne-cross/bin/linux-amd64/bartender

package: fyne-cross/bin/linux-amd64/bartender-v0.0.0-linux-amd64.tar.gz

fyne-cross/bin/linux-amd64/bartender:
	fyne-cross linux
	
fyne-cross/bin/linux-amd64/bartender-v0.0.0-linux-amd64.tar.gz:
	bash -c "cd fyne-cross/bin/linux-amd64 && tar $(basename $@) -zcvf bartender"

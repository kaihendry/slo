Version := $(shell git describe --tags --dirty)

amd64:
	docker build -t alexellis2/hendry:$@ . --build-arg TARGET_ARCH=$@
	docker push alexellis2/hendry:$@

arm:
	docker build -t alexellis2/hendry:$@ . --build-arg TARGET_ARCH=$@ --build-arg GOARM=6
	docker push alexellis2/hendry:$@

manifest: amd64 arm
	docker manifest create --amend alexellis2/hendry:$(Version) alexellis2/hendry:amd64 alexellis2/hendry:arm

	docker manifest annotate alexellis2/hendry:$(Version) alexellis2/hendry:arm --os linux --arch arm --variant v6
	
	docker manifest inspect alexellis2/hendry:$(Version)
	docker manifest push alexellis2/hendry:$(Version)

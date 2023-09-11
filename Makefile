.PHONY: init build

init:
	go mod init main
update:
	go mod tidy
build:
	make -f scripts/build.mk
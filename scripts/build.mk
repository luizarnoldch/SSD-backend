SHELL = cmd.exe
FOLDERS = get_historical_agent_metrics
GOOS = linux
GOARCH = amd64
CGO_ENABLED = 0
export GOOS GOARCH CGO_ENABLED

.PHONY: all $(FOLDERS)

all: $(FOLDERS)

$(FOLDERS):
	cmd /C "cd lambdas\$@ && \
	go build -o main && \
	build-lambda-zip.exe -o main.zip main && \
	move /Y main.zip ..\..\bin\$@.zip && \
	del /Q main "

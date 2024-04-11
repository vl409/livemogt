all: bin/livemogt bin/webmap

COMMON_SRCS=src/config.go src/daemon.go src/userinfo.go src/ringbuffer.go src/network.go

GO_ENV=CGO_ENABLED=0
GO_FLAGS=-ldflags '-s -w'

bin/livemogt: $(COMMON_SRCS) src/lmbot_gotelegram.go src/livemogt_msg.go src/livemogt.go
	$(GO_ENV) go build $(GO_FLAGS) -o $@ $^

bin/webmap: $(COMMON_SRCS) src/webmap.go
	$(GO_ENV) go build $(GO_FLAGS) -o $@ $^

clean:
	@rm -f bin/livemogt bin/webmap

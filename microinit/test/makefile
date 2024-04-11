all: test

.PHONY: test

test: payload
	MI_FOO=MI_BAR MI_BAR=MI_FOO \
                     ../microinit ./payload \
                                  ./payload:exit_1 \
                                  ./payload:exit_2 \
                                  ./payload:trap \
                                  ./payload:signal_int \
                                  ./payload:signal_kill \
                                  ./payload:a:bb:cccc:dddd \
                                  ./payload: \
                                  ./payload::

OPTS=-static -O2 -s

payload: payload.c
	$(CC) -Wall -Werror $(OPTS) -o $@ $^

clean:
	@rm -f payload

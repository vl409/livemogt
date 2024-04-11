all: microinit

OPTS=-static -O2 -s

microinit: microinit.c
	$(CC) -Wall -Werror $(OPTS) -o $@ $^

clean:
	@rm -f microinit

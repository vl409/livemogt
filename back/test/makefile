PROGS:=fake-users

all: $(PROGS)

fake-users: fake-users.go network.go
	go build -o $@ $^

clean:
	@rm -f $(PROGS)


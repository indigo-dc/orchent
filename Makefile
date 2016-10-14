
compile:
	./scripts/build.sh

install: compile
	install -m 755 orchent /usr/bin/orchent

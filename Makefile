compile_flags.txt:
	pkg-config --cflags libpipewire-0.3 | xargs -n1 echo > $@

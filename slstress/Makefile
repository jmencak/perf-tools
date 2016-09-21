PROGRAM=slstress
EXE=$(PROGRAM)$(EXESUF)

CC=gcc
#CC=diet gcc -s -O2

all: $(PROGRAM)$(EXESUF)
CCFLAGS=-Wall

$(EXE): $(PROGRAM).c
	$(CC) $(CCFLAGS) -o $(EXE) $(PROGRAM).c $(LIBS)

strip:
	strip $(PROGRAM)

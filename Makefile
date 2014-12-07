include $(GOROOT)/src/Make.inc

TARG=g/oscar

GOFILES= \
	reader.go \
	writer.go \
	client.go \
	family-04.go \
	family-13.go \
	family-15.go

include $(GOROOT)/src/Make.pkg

DATE ?= date
FIND ?= find
GIT ?= git
GO ?= go
GREP ?= grep
SED ?= sed
TOUCH ?= touch
TR ?= tr
UNAME ?= uname
XARGS ?= xargs
SHELL := /bin/bash

ifeq ($(shell uname), Darwin)
	SED = gsed
endif

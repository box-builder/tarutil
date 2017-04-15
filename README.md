# tarutil

[![GoDoc](https://godoc.org/github.com/box-builder/tarutil?status.svg)](https://godoc.org/github.com/box-builder/tarutil)
[![Build Status](http://jenkins.hollensbe.org:8080/buildStatus/icon?job=tarutil-master)](http://jenkins.hollensbe.org:8080/job/tarutil-master/)
[![Go Report Card](https://goreportcard.com/badge/github.com/box-builder/tarutil)](https://goreportcard.com/report/github.com/box-builder/tarutil)

This library is a collection of Go utilities to handle tar archives.

This library contains only code which can unpack tar archives with
AUFS whiteouts.

The layer unpacking code is based on Docker's pkg/archive.

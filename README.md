# tarutil

[![Build Status](http://jenkins.hollensbe.org:8080/buildStatus/icon?job=tarutil-master)](http://jenkins.hollensbe.org:8080/job/tarutil-master/)

This library is a collection of Go utilities to handle tar archives.

This library contains only code which can unpack tar archives with
AUFS whiteouts.

The layer unpacking code is based on Docker's pkg/archive.

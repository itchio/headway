# headway

[![Build Status](https://travis-ci.org/itchio/headway.svg?branch=master)](https://travis-ci.org/itchio/headway)
[![GoDoc](https://godoc.org/github.com/itchio/headway?status.svg)](https://godoc.org/github.com/itchio/headway)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/itchio/headway/blob/master/LICENSE)
[![No Maintenance Intended](http://unmaintained.tech/badge.svg)](http://unmaintained.tech/)

headway is a small Go library to track the progress of tasks.

**It is developed for internal use, I don't intend to take pull requests.**

It contains:

  * `ewma`: an exponential weighted moving average
  * `united`: formatting routines for bytes
  * `state`: a set of callbacks for log messages & progresses
  * `probar`: a CLI progress bar
  * `counter`: counting wrappers for `io.Reader` and `io.Writer`
  * `tracker`: a speed/ETA estimator for task progress


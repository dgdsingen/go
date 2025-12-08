package main

const maxLineLength = 64 * 1024 // 64KB
const appName = "r2n"

var version = "undefined"

var br = byte('\r')
var bn = byte('\n')

var bsr = []byte{br}
var bsn = []byte{bn}

// var bsnn = []byte{bn, bn}

package main

const maxLineLength = 64 << 10  // 64KB
const readBufLength = 32 << 10  // 32KB
const scanWindowLength = 1 << 7 // 128B
const appName = "r2n"

var version = "undefined"

var br = byte('\r')
var bn = byte('\n')

var bsr = []byte{br}
var bsn = []byte{bn}

// var bsnn = []byte{bn, bn}

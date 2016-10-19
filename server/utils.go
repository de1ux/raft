package main

import (
    "time"
    mathrand "math/rand"
	"net/http"
)

func rand(low, high int) int {
    mathrand.Seed(time.Now().UTC().UnixNano())
    if low == high {
        return -1
    }
    if low > high {
        oldLow := low
        low = high
        high = oldLow
    }
    normalizedHigh := high - low
    return mathrand.Intn(normalizedHigh) + low
}

func err500(w http.ResponseWriter, err error) {
	println(err.Error())
	w.WriteHeader(http.StatusInternalServerError)
}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}
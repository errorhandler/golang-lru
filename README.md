# golang-lru

This provides the `lru` package which implements a fixed-size
thread safe LRU cache. It is based on the cache in Groupcache.

It is a fork of the non-generic [golang-lru](https://github.com/hashicorp/golang-lru) from [Hashicorp](https://www.hashicorp.com/). The intetion is to track upstream closely, porting over generic typings.

# Documentation

Full docs are available on [Godoc](http://godoc.org/github.com/errorhandler/golang-lru)

# Example

Using the LRU is very simple:

```go
l, _ := New[int, int](128)
for i := 0; i < 256; i++ {
    l.Add(i, nil)
}
if l.Len() != 128 {
    panic(fmt.Sprintf("bad len: %v", l.Len()))
}
```

# Supported Go Versions

Following the [standard Go version support policy](https://go.dev/doc/devel/release#policy), this package will support the last two versions of Go that support generics.

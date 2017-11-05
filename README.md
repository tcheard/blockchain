# Basic Blockchain written in Go

Based on [Building Blockchain in Go](https://jeiwan.cc/posts/building-blockchain-in-go-part-1/) blog series.

## Building

Go 1.8+ is required to build blockchain, which uses the new vendor directory.

```
$ mkdir -p $GOPATH/src/github.com/tcheard
$ git clone https://github.com/tcheard/blockchain.git $GOPATH/src/github.com/tcheard/blockchain
$ cd $GOPATH/src/github.com/tcheard/blockchain
$ make
```

You should now have blockchain in your `bin/` directory:

```
$ ls bin/
blockchain
```
# see uniswap pending swaps in the txn pool!

press `h` for the help,

to start it, you'll need to connect with an endpoint, either IPC or websocket

I run my own custom geth node, so its default value is a local node, but you can pick something else

```
./swaps-tui -client_dial <ws://...>
```

# build

the usual go way, or invoke `make`

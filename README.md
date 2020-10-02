# see uniswap pending swaps in the txn pool!

see https://twitter.com/EdgarArout/status/1311908759250370560 for a gif of what it looks like

<video src="https://video.twimg.com/tweet_video/EjTVnChWAAAq6oB.mp4" preload="auto" playsinline="" type="video/mp4">
</video>

The bar charts show the value of the trade in eth

press `h` for the help,

to start it, you'll need to connect with an endpoint, either IPC or websocket

I run my own custom geth node, so its default value is a local node, but you can pick something else

```
./swaps-tui -client_dial <ws://...>
```

# build

the usual go way, or invoke `make`

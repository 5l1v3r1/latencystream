# Abstract

The **latencyreader** package makes it possible to read fixed-sized chunks from an `io.Reader` while staying below a maximum latency.

## Example use case

For example, suppose you are implementing a network protocol based on a carrier pigeon who can carry an 8GB USB stick between two locations. If you want to proxy a regular high-speed low-latency TCP connection through your pigeon protocol, it would be foolish to do one `read()` from the TCP socket, get roughly 64KB of data, and send it off by pigeon. This is clearly inefficient, because now you have to wait 10 minutes or so before your pigeon comes back. Instead, you should do enough `read()`s that you get 8GB of data from the socket, at which point you have no choice but to send that data off. However, this leaves you with one problem. Suppose the TCP connection goes idle after sending 4GB of data. After a fixed amount of time--say an hour--you should probably send off this data by pigeon anyway, even if you don't have a full 8GB.

# TODO

 * Remove all references to compile-time constants.
 * Wrap reader in a cancellable interface
 * Make everything else hidden

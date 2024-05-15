# go-libp2p-extras

Random tools for using go-libp2p. Nothing here is canon.

# Tools

`libp2pextras.WrapWithLogMissingSetDeadlines`: Wraps a host.Host and returns a version that will spy on streams to log if a `SetDeadline` is missing.

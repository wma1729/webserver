# A simple web server

Hash Requests are pushed on to a channel. The size of the channel is configurable via environment variable CHANNEL_SIZE. There are N worker threads waitingon the channel. The number of worker threads can be configured via environment variable NUM_OF_WORKERS.

On receiving the hash request, a unique id is generated and sent back to the client and the request is pushed into the channel.
The worker threads pick up the request from the channel, wait if needed, and then compute the hashed password. A map of request ID -> hashed password is maintained.

On receiving the lookup request, the map is used to fulfil the request.

A global ServerStatistics is maintained to keep track of the number of requests received for stats endpoint.

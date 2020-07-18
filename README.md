# goChat
Small console chat written with Golang

## architecture
1) we have a listening TCP socket + N connection sockets
2) each connection socket is served by goroutine
3) when we get a message from client, we should broadcast it to all other clients
4) when we get a new connection from a client, we prepare and fill apopriate structure per client (nickname, channel, userid)
5) every goroutine writes data to all channels exept its own and listen to messages to its own channel
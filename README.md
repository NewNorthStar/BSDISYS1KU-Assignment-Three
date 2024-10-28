# Chitty-Chat System
Assignment hand in for "Distributed Systems, BSc (Autumn 2024) / Mandatory Activity 3" - A chat service implemented in Go! using gRPC. 

## Authors
- ghej@itu.dk
- rono@itu.dk
- luvr@itu.dk

## How to Run
1. Run `server.go` from command line. The server will listen for participants on `localhost:5050`. 
2. In other command line windows, run `client.go` as chat participants. These connect to `localhost:5050`. You are asked to type in a _callsign_, on startup, and this will be the name of the participant for the session. 
3. Participants can post messages by typing them in terminal and hitting `ENTER`. The server will disconnect a participant sending a message that is too long. (Maximum is 128 utf-8 characters.) 
4. To disconnect as a participant, simply terminate the program. Normally `Ctrl+C`. 
5. Stopping the server disconnects all participants. 
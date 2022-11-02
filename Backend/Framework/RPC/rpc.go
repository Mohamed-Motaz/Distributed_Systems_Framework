package RPC

/*
This package contains all the RPC definitions
for any inter-servers communication
*/

//master-worker communication ---------
type SayHello struct {
	Content string
}

//master-lockserver communication ---------

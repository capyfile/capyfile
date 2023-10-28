package main

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	server := &Server{
		Addr: ":8024",
	}

	serverInitErr := server.Init()
	if serverInitErr != nil {
		panic(serverInitErr)
	}

	serverRunErr := server.Run()
	if serverRunErr != nil {
		panic(serverRunErr)
	}
}

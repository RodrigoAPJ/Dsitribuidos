DN: clean
	go run ./datanode/datanode.go

NN: clean
	go run ./namenode/namenode.go

Pozo: clean
	go run ./pozo/pozo.go

Leader: clean
	go run ./leader/leader.go

Players: clean
	go run ./players/players_controller.go

clean: 
	rm -f ./namenode/*.txt
	rm -f ./datanode/DN1/*.txt
	rm -f ./datanode/DN2/*.txt
	rm -f ./datanode/DN3/*.txt
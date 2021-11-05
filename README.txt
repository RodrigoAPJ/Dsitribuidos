Tras modificar archivo squidgame.proto perteneciente a squidgame/grpc aplicar comandos:
	export PATH="$PATH:$(go env GOPATH)/bin"
	protoc -I ./ --go_out=./ --go-grpc_out=./ ./squidgame.proto
(OJO: aplicarlos dentro de la misma carpeta squidgame/grpc)

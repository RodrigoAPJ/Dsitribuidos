Nombres:
Fabian Jorquera 201573513-1
Daniel Magaña 201873504-3
Rodrigo Pérez 201873619-8

Instrucciones:

Debera correr en el siguiente orden los comandos obligatoriamente, en caso contrario fallará:

Abrir máquina 117:
	estando en dentro de carpeta SquidGame:	
		make DN 	//Ese comando iniciara un archivo datanode.go y le preguntara por el nombre de la máquina, ahí debe responder 117 para que así el archivo sepa que ip y puerto abrir.

Abrir máquina 119:
	estando en dentro de carpeta SquidGame:	
		make DN 	//Se le preguntara por el nombre de la máquina, ahí debe responder 119

Abrir máquina 120:
	estando en dentro de carpeta SquidGame:	
		make DN 	//Se le preguntara por el nombre de la máquina, ahí debe responder 120

Abrir máquina 118:
	estando en dentro de carpeta SquidGame:
		make NN 	//Esto iniciara el servidor namenode que se conectara automáticamente a los DataNodes de las otras máquinas

Volver a la máquina 117:
	Apretrar CTRL+Z y luego escribir bg, así el dataNode seguira ejecutandose por detras y podra iniciar el Pozo.

	make Pozo	//Iniciara el servidor GRPC del Pozo

Volver a la máquina 119:
	Apretrar CTRL+Z y luego escribir bg, así el dataNode seguira ejecutandose por detras y podra iniciar al Lider.

	make Leader		//Iniciara el servidor GRPC del Lider

Volver a la máquina 120:
	Apretrar CTRL+Z y luego escribir bg, así el dataNode seguira ejecutandose por detras y podra iniciar a los jugadores.

	make Players	//Iniciara a los jugadores


Si todo salio bien podra ver en consola muchos printeos indicando que tal proceso se conecto a tal servidor en las distintas consolas. Además los jugadores se conectaran automáticamente al servidor Leader (por lo que debería printearse que 16 jugadores se conectaron en la consola del Lider)

POSDATA: Durante el primer juego en la consola de los jugadores se le preguntara si quiere que los bots hagan jugadas inteligentes para aumentar sus probabilidades de sobrevivir, si opta por no aceptar jugaran normal y por azar morira la mayoria dado que es dificil por probabilidad que sobrevivan muchos...

Una vez que termine de probar dando los inputs por consola y tal podra ver en la carpeta namenode de la máquina 118 un archivo llamado Registro.txt que tiene escrito en que servidor se guardo la jugada de cada jugador.

A su vez en la carpeta datanode de la máquina 117, 119 y 120 encontrará 3 carpetas DN1, DN2, DN3: 
	si se encuentra en la máquina 117 encontrará dentro de la carpeta DN1 todos los archivos con las jugadas de todos los jugadores que por azar el nameNode eligio guardar en este dataNode

	si se encuentra en la máquina 119 encontrará dentro de la carpeta DN2 todos los archivos con las jugadas de todos los jugadores que por azar el nameNode eligio guardar en este dataNode

	si se encuentra en la máquina 120 encontrará dentro de la carpeta DN3 todos los archivos con las jugadas de todos los jugadores que por azar el nameNode eligio guardar en este dataNode


SI QUIERE VOLVER A CORRER TODO DENUEVO RECUERDE MATAR LOS PROCESOS QUE ESTAN CORRIENDO EN EL BACKGROUND, PARA ELLO VAYA A LAS CONSOLAS Y HAGA LO SIGUIENTE:

	jobs		//Para verificar que hay procesos corriendo en background

	kill %1 	//Para matar el proceso corriendo en background

UNA VEZ HAGA ESTO EN TODAS LAS CONSOLAS CON PROCESOS CORRIENDO EN BACKGROUND PRODRA COMENZAR A ESCRIBIR TODOS LOS COMANDOS DE MAKE DENUEVO (OJO QUE ESTOS BORRARAN LOS REGISTROS DE NAMENODE Y DATANODES ANTERIORES ANTES DE COMENZAR)


COSAS A TENER EN CONSIDERACIÓN:

Si bien son 16 jugadores, estos se identifican desde el 0 hasta el 15, además nosotros somos el jugador 15


Se manejo a los jugadores de la siguiente forma:
	Jugador elije por consola su acción (jugar juego 1,2 o 3), y luego se generan las jugadas de los bots para ese juego, si el servidor aún no inicia tal etapa de juego entonces las jugadas no se procesaran y en la consola del jugador se podra observar que no se procesaron las jugadas.
	Si el jugador muere en el primer Juego seguira actuando como controlador, osea, podra elejir la siguiente accion (jugar juego 1, 2 o 3) pero el no podra jugar, solo los bots jugaran el juego (y de todas maneras se verifica si el servidor esta jugando tal juego para así procesar o no la jugada)... DEBIDO A ESTO NO SE CUMPLE CON "El proceso del jugador termina solo al ganar el juego del calamar o al ser eliminado" ... lamentablemente vimos la pauta despúes y ya habiamos avanzado mucho la tarea como para modificar nuestra implementación :(

NO ALCANZAMOS A IMPLEMENTAR EL JUEGO NÚMERO 3

NINGÚN PROCESO DE TIPO SERVIDOR SE PUEDE TERMINAR POR CONSOLA, SOLAMENTE USANDO CTRL+C... OJO QUE LA OPCIÓN DE LOS JUGADORES PARA SALIR SI FUNCIONA DADO QUE NO ES SERVIDOR. ESO SI, SI APRIETA EN LA CONSOLA DEL LIDER LA OPCION 4 PARA SALIR SE PRINTEARA SI ES QUE HUBO GANADORES O NO.

POSDATA: SI POR ALGUNA RAZON TIRA FALLO O ALGO ANDA RARO AVISAR PARA VER EN RECORRECCIÓN PORQUE YA LO PROBAMOS POR PIES Y CABEZA Y NO DEBERIA PORQUE FUNCIONAR MAL...
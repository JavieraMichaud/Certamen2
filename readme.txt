Simulador de Planificación de Procesos
----------------------------------------

Este proyecto contiene un simulador de planificación de procesos escrito en Go.

Requisitos 

- Go instalado en el sistema.  

Estructura del Proyecto 

- main.go: 
  Código fuente Go que implementa la lógica de simulación. Define las estructuras de datos para representar procesos, colas, dispatcher. También contiene las funciones para crear procesos, leer archivos de entrada, ejecutar instrucciones, etc.

- Creacion_Procesos.txt:
  Archivo de entrada que define las órdenes de creación de los procesos a simular.

- Carpeta Procesos/: 
  Contiene archivos .txt con las instrucciones de cada proceso específico.

- archivo_salida.txt:
  Archivo donde se escribe la salida de la simulación.

-dispatcher.go: 
    Implementación del despachador y las funciones relacionadas con la ejecución y cambio de procesos.

- proceso.go: 
    Definición de la estructura del proceso y funciones relacionadas.

- cola.go: 
    Definición de la estructura de la cola y funciones asociadas.

- leerOrdenes.go: 
    Funciones para leer las órdenes de creación de procesos desde un archivo.

Uso

Para ejecutar la simulación:

go run main.go n o p carpeta_procesos.txt nombre_archivo_salida.txt

Donde:
  - n = cantidad de instrucciones a ejecutar por ciclo 
  - o = quantum 
  - p = probabilidad de finalización anticipada
  - carpeta_procesos = Carpeta que contiene archivos de procesos.
  - nombre_archivo_salida = Nombre del archivo de salida para registrar los resultados de la simulación.
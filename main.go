package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// OrdenCreacion representa la orden de creación de un proceso.
type OrdenCreacion struct {
	Tiempo  int
	Archivo string
}

// Proceso representa un proceso en el sistema.
// La estructura contiene información sobre el ID, nombre, estado, contador de instrucciones,
// conjunto de instrucciones, estado de bloqueo, estado de finalización, un mutex para sincronización y
// un puntero al siguiente proceso en una cola.
type Proceso struct {
	ID            int        // Identificador único del proceso
	Nombre        string     // Nombre del proceso
	Estado        string     // Estado actual del proceso (Nuevo, Listo, Ejecutando, Bloqueado, Finalizado)
	Contador      int        // Contador de instrucciones ejecutadas por el proceso
	Instrucciones []string   // Conjunto de instrucciones que el proceso debe ejecutar
	EsBloqueado   bool       // Indica si el proceso está bloqueado por una operación de E/S
	EsFinalizado  bool       // Indica si el proceso ha finalizado su ejecución
	Mutex         sync.Mutex // Mutex utilizado para garantizar la sincronización de datos del proceso
	siguiente     *Proceso   // Puntero al siguiente proceso en la cola (si es parte de una cola)
}

// Cola representa una cola de procesos.
type Cola struct {
	PrimerProceso *Proceso
	UltimoProceso *Proceso
}

// Dispatcher maneja la ejecución y cambio de procesos.
type Dispatcher struct {
	ProcesoActual         *Proceso
	ColaListo             *Cola
	ColaBloqueado         *Cola
	ContadorInstrucciones int
}

// Estados posibles de un proceso.
const (
	EstadoNuevo      = "Nuevo"
	EstadoListo      = "Listo"
	EstadoEjecutando = "Ejecutando"
	EstadoBloqueado  = "Bloqueado"
	EstadoFinalizado = "Finalizado"
)

var procesosDesalojados int
var quantum int
var instruccion string
var contadorProcesos int

func main() {
	// Validar la cantidad correcta de argumentos
	if len(os.Args) != 6 {
		fmt.Println("Uso: go run main.go n o p carpeta_procesos nombre_archivo_salida")
		return
	}

	// Validar y convertir n, o = quantum y p a enteros

	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n <= 0 {
		fmt.Println("Error: n debe ser un número entero positivo")
		return
	}

	quantum, err := strconv.Atoi(os.Args[2])
	if err != nil || quantum <= 0 {
		fmt.Println("Error: o debe ser un número entero positivo")
		return
	}

	p, err := strconv.Atoi(os.Args[3])
	if err != nil || p <= 0 {
		fmt.Println("Error: p debe ser un número entero positivo")
		return
	}

	// archivoSalida almacena el nombre del archivo de salida especificado como argumento de línea de comandos.
	// Este archivo se utilizará para escribir la salida de la simulación.
	archivoSalida := os.Args[5]

	// Dispatcher maneja la ejecución y cambio de procesos en la simulación.
	dispatcher := &Dispatcher{
		ColaListo:     &Cola{}, // Inicializar cola de procesos listos
		ColaBloqueado: &Cola{}, // Inicializar cola de procesos bloqueados
	}

	// Leer archivos de entrada
	ordenes, err := leerOrdenes("Creacion_Procesos.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Cargar procesos según órdenes
	for _, o := range ordenes {
		archivoProceso := filepath.Join("Procesos", o.Archivo+".txt")
		fmt.Println("Intentando abrir el archivo:", archivoProceso)
		proceso, err := cargarProcesoDesdeArchivo(archivoProceso)
		if err != nil {
			fmt.Printf("Error al cargar proceso desde archivo %s: %v\n", archivoProceso, err)
			return
		}
		agregarProcesoACola(dispatcher.ColaListo, proceso) //agrega un proceso a la cola de procesos listos
	}
	// Ejecutar la simulación con los parámetros proporcionados
	ejecutarSimulacion(dispatcher, n, quantum, p, archivoSalida)
}

// Función principal para ejecutar la simulación.
func ejecutarSimulacion(dispatcher *Dispatcher, n, q, p int, archivoSalida string) {
	// Crear un archivo de salida para escribir los resultados de la simulación.
	file, err := os.Create(archivoSalida)
	if err != nil {
		// Manejar el error si no se puede crear el archivo de salida.
		fmt.Println("Error al crear archivo de salida:", err)
		return
	}
	defer file.Close() // Cerrar el archivo al finalizar la función.

	// Ejecutar la simulación mientras haya un proceso actual en el Dispatcher.
	for dispatcher.ProcesoActual != nil {
		ejecutarInstrucciones(dispatcher, n, q, p) // Ejecutar las instrucciones del proceso actual

		// Generar una línea de trama con la información actual y escribirla en el archivo de salida
		lineaTrama := generarLineaTrama(dispatcher)
		fmt.Fprintln(file, lineaTrama)

		// Cambiar al siguiente proceso en la cola de procesos listos
		cambiarProceso(dispatcher)
	}
	// Finalizar la simulación después de que no hay más procesos para ejecutar.
	finalizar(dispatcher)
}

// Función para ejecutar las instrucciones de un proceso.
// ejecutarInstrucciones ejecuta las instrucciones de un proceso actual en el Dispatcher.
// n representa la cantidad de instrucciones a ejecutar en un ciclo.
// q representa la longitud del quantum, y p es la probabilidad de terminación anticipada.
func ejecutarInstrucciones(dispatcher *Dispatcher, n, q, p int) {
	for i := 0; i < n; i++ {
		// Bloquear el mutex del proceso actual para acceder a sus datos de manera segura
		dispatcher.ProcesoActual.Mutex.Lock()

		// Obtener la instrucción actual del proceso
		instruccion = dispatcher.ProcesoActual.Instrucciones[dispatcher.ProcesoActual.Contador]

		// Desbloquear el mutex después de obtener la instrucción
		dispatcher.ProcesoActual.Mutex.Unlock()

		// Procesar la instrucción actual
		if instruccion == "I" {
			// Incrementar el contador de instrucciones y marcar como finalizado si es necesario
			dispatcher.ProcesoActual.Contador++
			if dispatcher.ProcesoActual.Contador == len(dispatcher.ProcesoActual.Instrucciones) {
				dispatcher.ProcesoActual.EsFinalizado = true
			}
		} else if instruccion == "ES" {
			// Incrementar el contador y marcar como bloqueado
			dispatcher.ProcesoActual.Contador++
			dispatcher.ProcesoActual.EsBloqueado = true

			// Mostrar evento E/S y agregar a la cola de procesos bloqueados
			fmt.Println("Tiempo de CPU", dispatcher.ContadorInstrucciones, "Evento E/S", dispatcher.ProcesoActual.Nombre)
			agregarProcesoACola(dispatcher.ColaBloqueado, dispatcher.ProcesoActual)
		} else if instruccion == "F" && rand.Float64() < 1/float64(p) {
			// Simular terminación anticipada con probabilidad 1/p
			dispatcher.ProcesoActual.EsFinalizado = true
		}

		// Incrementar el contador global de instrucciones ejecutadas
		dispatcher.ContadorInstrucciones++

		// Verificar si se ha alcanzado el límite del quantum
		if dispatcher.ProcesoActual.Contador >= quantum {
			// Reiniciar el contador del quantum, incrementar procesos desalojados y salir del ciclo
			dispatcher.ProcesoActual.Contador = 0
			procesosDesalojados++
			return
		}
	}
}

// Cambiar al siguiente proceso en la cola de listos.
func cambiarProceso(dispatcher *Dispatcher) {
	// Bloquear el mutex del proceso actual para realizar cambios de manera segura
	dispatcher.ProcesoActual.Mutex.Lock()
	defer dispatcher.ProcesoActual.Mutex.Unlock()

	// Verificar si hay un proceso actual
	if dispatcher.ProcesoActual != nil {
		// Reiniciar el contador del proceso actual y agregarlo a la cola de listos
		dispatcher.ProcesoActual.Contador = 0
		agregarProcesoACola(dispatcher.ColaListo, dispatcher.ProcesoActual)
	}

	// Obtener el siguiente proceso de la cola de listos
	dispatcher.ProcesoActual = quitarProcesoDeCola(dispatcher.ColaListo)
	if dispatcher.ProcesoActual == nil {
		return
	}

	// Mostrar información sobre el cambio de proceso en la salida estándar
	fmt.Printf("Tiempo de CPU %d Tipo Instrucción %s Proceso %s Valor CP %d\n",
		dispatcher.ContadorInstrucciones, instruccion, dispatcher.ProcesoActual.Nombre, dispatcher.ProcesoActual.Contador)

	// Verificar si el nuevo proceso supera el límite del quantum
	if dispatcher.ProcesoActual.Contador >= quantum {
		// Incrementar la cantidad de procesos desalojados y reiniciar el contador del proceso
		procesosDesalojados++
		dispatcher.ProcesoActual.Contador = 0
		// Obtener el siguiente proceso de la cola de listos
		dispatcher.ProcesoActual = quitarProcesoDeCola(dispatcher.ColaListo)
	}
}

// Finalizar la simulación y mostrar el número de procesos desalojados
func finalizar(dispatcher *Dispatcher) {
	// Mostrar la cantidad de procesos desalojados al finalizar la simulación
	fmt.Println(procesosDesalojados)
}

// Generar una línea de trama para escribir en el archivo de salida.
func generarLineaTrama(dispatcher *Dispatcher) string {
	// Bloquear el mutex del proceso actual para obtener información de manera segura
	dispatcher.ProcesoActual.Mutex.Lock()

	// Utilizar "defer" para desbloquear el mutex asociado al proceso actual una vez que la función que contiene esta línea haya finalizado.
	defer dispatcher.ProcesoActual.Mutex.Unlock()

	// Crear una línea de trama con información sobre el estado actual del proceso
	linea := fmt.Sprintf("%d %s %s %d",
		dispatcher.ContadorInstrucciones,
		instruccion,
		dispatcher.ProcesoActual.Nombre,
		dispatcher.ProcesoActual.Contador)

	// Agregar detalles adicionales si el proceso no ha finalizado
	if dispatcher.ProcesoActual.Contador < len(dispatcher.ProcesoActual.Instrucciones) {
		linea += " " + dispatcher.ProcesoActual.Instrucciones[dispatcher.ProcesoActual.Contador]
	}

	return linea
}

// Crear un nuevo proceso a partir de un conjunto de instrucciones.
func crearProceso(instrucciones []string) *Proceso {
	// Verificar si es el primer proceso (despachador) y asignar un ID especial
	if contadorProcesos == 0 {
		return &Proceso{
			ID:            100,
			Nombre:        fmt.Sprintf("Proceso_%d", 100),
			Estado:        EstadoNuevo,
			Contador:      0,
			Instrucciones: instrucciones,
			EsBloqueado:   false,
			EsFinalizado:  false,
		}
	} else {
		// Crear un nuevo proceso con un ID único
		return &Proceso{
			ID:            contadorProcesos,
			Nombre:        fmt.Sprintf("Proceso_%d", contadorProcesos),
			Estado:        EstadoNuevo,
			Contador:      0,
			Instrucciones: instrucciones,
			EsBloqueado:   false,
			EsFinalizado:  false,
		}
	}
}

// Agregar un proceso a la cola.
func agregarProcesoACola(cola *Cola, proceso *Proceso) {
	// Verificar si la cola está vacía y agregar el proceso como el primer elemento
	if cola.PrimerProceso == nil {
		cola.PrimerProceso = proceso
	} else {
		// Agregar el proceso al final de la cola utilizando el campo siguiente
		cola.UltimoProceso.siguiente = proceso
	}

	// Actualizar el último proceso en la cola
	cola.UltimoProceso = proceso
}

// Quitar un proceso de la cola.
func quitarProcesoDeCola(cola *Cola) *Proceso {
	// Verificar si la cola está vacía
	if cola.PrimerProceso == nil {
		return nil
	}

	// Obtener el primer proceso de la cola
	proceso := cola.PrimerProceso

	// Actualizar el primer proceso en la cola utilizando el campo siguiente
	cola.PrimerProceso = cola.PrimerProceso.siguiente

	// Devolver el proceso obtenido de la cola
	return proceso
}

// Leer las órdenes de creación de procesos desde un archivo.
// La función toma el nombre del archivo como parámetro y devuelve una lista de estructuras OrdenCreacion y un posible error.
func leerOrdenes(archivo string) ([]OrdenCreacion, error) {
	// Abrir el archivo para lectura
	f, err := os.Open(archivo)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Inicializar la lista de órdenes de creación
	var ordenes []OrdenCreacion

	// Crear un scanner para leer el archivo línea por línea
	scanner := bufio.NewScanner(f)

	// Iterar sobre cada línea del archivo
	for scanner.Scan() {
		// Obtener el contenido de la línea
		linea := scanner.Text()

		// Ignorar líneas que comienzan con "#" (comentarios)
		if strings.HasPrefix(linea, "#") {
			continue
		}

		// Dividir la línea en campos utilizando espacios en blanco
		campos := strings.Fields(linea)

		// Ignorar líneas con menos de dos campos
		if len(campos) < 2 {
			continue
		}

		// Convertir el primer campo a un número entero (tiempo)
		tiempo, err := strconv.Atoi(campos[0])
		if err != nil {
			continue
		}

		// Verificar si es la primera orden de creación (proceso despachador)**
		if contadorProcesos == 0 {
			// Crear una orden de creación especial para el proceso despachador
			orden := OrdenCreacion{
				Tiempo:  tiempo,
				Archivo: "despachador",
			}

			// Agregar la orden a la lista de órdenes
			ordenes = append(ordenes, orden)

			// Incrementar el contador de procesos
			contadorProcesos++
		} else {
			// Crear una orden de creación normal utilizando el segundo campo como nombre de archivo
			orden := OrdenCreacion{
				Tiempo:  tiempo,
				Archivo: campos[1],
			}

			// Agregar la orden a la lista de órdenes
			ordenes = append(ordenes, orden)
		}
	}

	// Verificar si hubo errores durante la lectura del archivo
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Devolver la lista de órdenes de creación y nil para el error
	return ordenes, nil
}

// cargarProcesoDesdeArchivo carga un proceso desde un archivo dado.
// Lee las instrucciones desde el archivo y crea un nuevo proceso con ellas.
// Devuelve el proceso creado o un error si el archivo está vacío.
func cargarProcesoDesdeArchivo(archivo string) (*Proceso, error) {
	file, err := os.Open(archivo)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Variable para almacenar las instrucciones del proceso
	var instrucciones []string

	// Escanear el archivo línea por línea
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Obtener la línea actual
		linea := scanner.Text()

		// Ignorar líneas de comentarios y líneas vacía
		if strings.HasPrefix(linea, "#") || len(linea) == 0 {
			continue
		}

		// Agregar la instrucción a la lista
		instrucciones = append(instrucciones, strings.TrimSpace(linea))
	}
	// Verificar si hay instrucciones en el archivo
	if len(instrucciones) > 0 {
		// Crear un nuevo proceso con las instrucciones
		return crearProceso(instrucciones), nil
	}
	// Devolver un error si el archivo de proceso está vacío
	return nil, errors.New("el archivo de proceso está vacío")
}

package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Empty struct{}

type info struct {
	id    int
	tipus int
	nom   string
	salo  int
}

const FUMADORS = 6
const NoFUMADORS = 6

var (
	salons [3]int
	taules [3]int

	nom = []string{"Miquel", "Nati", "Pilar", "Albert",
		"Elna", "Ariadna", "Abdona", "Selma", "Anscari", "Darius", "Emilia", "Adali"}
	//Canales que representan los permisos
	demanaTaula  = make(chan info)
	demanaCompte = make(chan info)
	permisTaula  [FUMADORS + NoFUMADORS]chan int
	permisCompte = make(chan Empty)
)

func getDisponibilidad(tipus int) int {
	for i := 0; i < len(salons); i++ {
		if salons[i] == -1 || salons[i] == tipus {
			if taules[i] > 0 {
				return i
			}
		}
	}
	return -1
}

func getComensals(salo int) int {
	return (3 - taules[salo])
}

func noFumador(nom string, id int, done chan Empty) {
	fmt.Printf("Hola, el meu nom és %s, voldria dinar i som no fumador/a\n", nom)
	var mevaInfo info
	mevaInfo.id = id
	mevaInfo.nom = nom
	mevaInfo.tipus = 0

	demanaTaula <- mevaInfo
	salo := <-permisTaula[id]
	mevaInfo.salo = salo
	fmt.Printf("%s diu: M'agrada molt el saló %d\n", nom, salo)
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
	fmt.Printf("%s diu: Ja he dinat, el compte per favor\n", nom)
	demanaCompte <- mevaInfo
	<-permisCompte
	done <- Empty{}
}

func fumador(nom string, id int, done chan Empty) {
	fmt.Printf("Hola, el meu nom és %s, voldria dinar i som fumador/a\n", nom)
	var mevaInfo info
	mevaInfo.id = id
	mevaInfo.nom = nom
	mevaInfo.tipus = 1

	demanaTaula <- mevaInfo
	salo := <-permisTaula[id]
	mevaInfo.salo = salo
	fmt.Printf("%s diu: M'agrada molt el saló %d\n", nom, salo)
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
	fmt.Printf("%s diu: Ja he dinat, el compte per favor\n", nom)
	demanaCompte <- mevaInfo
	<-permisCompte
	done <- Empty{}
}

func proveedor() {
	esperenNoFumadors := [FUMADORS]int{0, 0, 0, 0, 0, 0}
	esperenFumadors := [NoFUMADORS]int{0, 0, 0, 0, 0, 0}
	numEsperenNoFumadors := 0
	numEsperenFumadors := 0
	var salo int
	for {
		select {
		case info := <-demanaTaula:
			if info.tipus == 0 {
				if getDisponibilidad(0) != -1 {
					salo = getDisponibilidad(0)
					permisTaula[info.id] <- salo
					fmt.Printf("***** El Sr./Sra. %s té taula al saló %d de NOFUMADORs\n", info.nom, salo)
					if taules[salo] == 3 {
						salons[salo] = info.tipus
					}
					taules[salo] = taules[salo] - 1
				} else {
					fmt.Printf("No hi ha cap taula lliure per a %s a NOFUMADOR\n", info.nom)
					esperenNoFumadors[info.id] = 1
					numEsperenNoFumadors++
				}
			}
			if info.tipus == 1 {
				if getDisponibilidad(1) != -1 {
					salo = getDisponibilidad(1)
					permisTaula[info.id] <- salo
					fmt.Printf("***** El Sr./Sra. %s té taula al saló %d de FUMADORs\n", info.nom, salo)
					if taules[salo] == 3 {
						salons[salo] = info.tipus
					}
					taules[salo] = taules[salo] - 1
				} else {
					fmt.Printf("No hi ha cap taula lliure per a %s a FUMADOR\n", info.nom)
					esperenFumadors[info.id] = 1
					numEsperenFumadors++
				}
			}
		case info := <-demanaCompte:
			if info.tipus == 0 {
				taules[info.salo] = taules[info.salo] + 1
				if taules[info.salo] == 3 {
					salons[info.salo] = -1
				}
				permisCompte <- Empty{}
				fmt.Printf("***** S'allibera un lloc del saló %d NOFUMADOR. Queden %d comensals\n", info.salo, getComensals(info.salo))
				if numEsperenNoFumadors > 0 {
					i := 0
					for i = 0; i < (FUMADORS); i++ {
						if esperenNoFumadors[i] == 1 {
							break
						}
					}
					taules[info.salo] = taules[info.salo] - 1
					fmt.Printf("***** El Sr./Sra. %s té taula al saló %d de NOFUMADORs\n", nom[i], info.salo)
					permisTaula[i] <- info.salo
					esperenNoFumadors[i] = 0
					numEsperenNoFumadors--
				}
			}
			if info.tipus == 1 {
				taules[info.salo] = taules[info.salo] + 1
				if taules[info.salo] == 3 {
					salons[info.salo] = -1
				}
				permisCompte <- Empty{}
				fmt.Printf("***** S'allibera un lloc del saló %d FUMADOR. Queden %d comensals\n", info.salo, getComensals(info.salo))
				if numEsperenFumadors > 0 {
					i := 0
					for i = 0; i < (NoFUMADORS); i++ {
						if esperenFumadors[i] == 1 {
							break
						}
					}
					taules[info.salo] = taules[info.salo] - 1
					fmt.Printf("***** El Sr./Sra. %s té taula al saló %d de FUMADORs\n", nom[i+6], info.salo)
					permisTaula[i] <- info.salo
					esperenFumadors[i] = 0
					numEsperenFumadors--
				}
			}
		}
	}
}

func main() {
	for i := 0; i < len(salons); i++ { // Al principio los salones no están restringidos
		salons[i] = -1
	}
	for i := 0; i < len(taules); i++ { // Al principio todos los salones tienen 3 mesas libres
		taules[i] = 3
	}

	for i := range permisTaula {
		permisTaula[i] = make(chan int)
	}
	done := make(chan Empty)

	go proveedor()

	for i := 0; i < NoFUMADORS; i++ {
		go noFumador(nom[i], i, done)
	}
	for i := 0; i < FUMADORS; i++ {
		go fumador(nom[i+6], i, done)
	}

	for i := 0; i < FUMADORS; i++ {
		<-done
	}
	for i := 0; i < NoFUMADORS; i++ {
		<-done
	}

	//<-done

	fmt.Println("SIMULACIÓ ACABADA")

}

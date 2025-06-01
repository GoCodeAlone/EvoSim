package main

import (
	"fmt"
	"log"
)

// RunWebInterface starts the web interface for the simulation
func RunWebInterface(world *World, port int) error {
	// TODO: Implement web interface with websockets
	// This is a placeholder to make the code compile
	
	fmt.Printf("Web interface would start on port %d\n", port)
	fmt.Println("Web interface not yet implemented")
	log.Println("Use CLI mode for now: run without --web flag")
	
	return fmt.Errorf("web interface not yet implemented")
}
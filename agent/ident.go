package agent

import (
	"bufio"
	"log"
	"os"

	satori "github.com/satori/go.uuid"
)

func CreateOrGetAgentId(idFile string) string {
	if _, err := os.Stat(idFile); os.IsNotExist(err) {
		agentId := getUUID()
		log.Println("Created agentId", agentId)
		writeAgentId(agentId, idFile)
		return agentId
	}
	return readAgentId(idFile)
}

func getUUID() string {
	uuid, err := satori.NewV4()
	if err != nil {
		panic(err)
	}
	return uuid.String()
}

func writeAgentId(agentId, idFile string) {
	file, err := os.OpenFile(idFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = file.WriteString(agentId + "\n")
}

func readAgentId(idFile string) string {
	file, err := os.Open(idFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	agentId := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		agentId = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
	log.Println("Using agentId", agentId)
	return agentId
}

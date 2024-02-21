package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Historico struct {
	Pergunta string `json:"pergunta"`
	Resposta string `json:"resposta"`
}

var historico []Historico
var nomeArquivo = "historico.json"

func main() {
	carregarHistorico()

	gerarReceita()
}

func carregarHistorico() {
	if _, err := os.Stat(nomeArquivo); err == nil {
		conteudo, err := os.ReadFile(nomeArquivo)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(conteudo, &historico)
	}
}

func salvarHistorico() {
	conteudo, err := json.Marshal(historico)
	if err != nil {
		panic(err)
	}
	os.WriteFile(nomeArquivo, conteudo, 0644)
}

func gerarReceita() {
	reader := bufio.NewReader(os.Stdin)

	var mensagem string
	fmt.Println("Do que você quer a receita:\n\nDigite ou 'sair' para encerrar:")

	mensagem, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Erro ao ler entrada:", err)
		return
	}

	if mensagem == "sair" {
		return
	} else if mensagem == "limpar" {
		historico = []Historico{}
		salvarHistorico()
		fmt.Println("Histórico limpo :)")
		gerarReceita()
		return
	}

	// https://platform.openai.com/account/billing
	// https://platform.openai.com/api-keys - gerar tokens

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY não definida.")
		return
	}

	messages := []map[string]string{
		{
			"role":    "system",
			"content": "Você é um assistente virtual. e tem o objetivo em criar curriculos",
		},
	}

	for _, h := range historico {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": h.Pergunta,
		})
		messages = append(messages, map[string]string{
			"role":    "assistant",
			"content": h.Resposta,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": mensagem,
	})

	url := "https://api.openai.com/v1/chat/completions"
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":    "gpt-3.5-turbo", // Substitua pelo modelo desejado.
		"messages": messages,
	})

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Use io.ReadAll aqui para ler a resposta
	responseBody, erro := io.ReadAll(resp.Body)

	if erro != nil {
		fmt.Println(err.Error())
		return
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(responseBody), &result)

	choices := result["choices"].([]interface{})
	firstChoice := choices[0].(map[string]interface{})
	message := firstChoice["message"].(map[string]interface{})
	resposta := message["content"].(string)

	registraEMostraResposta(mensagem, resposta)
}

func registraEMostraResposta(mensagem, resposta string) {
	historico = append(historico, Historico{Pergunta: mensagem, Resposta: resposta})
	salvarHistorico()
	fmt.Printf("%s\n\nOu digite 'sair' para encerrar\nDigite ...", resposta)
	gerarReceita()
}

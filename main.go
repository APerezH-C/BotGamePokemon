package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "github.com/google/uuid"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	productURL    = "https://www.game.es/ACCESORIOS/CONTROLLER/PLAYSTATION-5/MANDO-INALAMBRICO-DUALSENSE-BLANCO-V2/225820"
	checkInterval = 60 * time.Second // Chequea cada 60 segundos
)

var (
	telegramToken  = os.Getenv("TELEGRAM_TOKEN")
	telegramChatID = os.Getenv("TELEGRAM_CHAT_ID")
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Mobile/15E148 Safari/604.1",
}

func main() {
	for {
		fmt.Println("🔍 Revisando stock...")

		inStock, err := checkStock()
		if err != nil {
			fmt.Println("⚠️ Error revisando el stock:", err)
		} else if inStock {
			fmt.Println("🟢 ¡HAY STOCK! Visita:", productURL)

			msg := fmt.Sprintf("🟢 ¡El producto está disponible!\n\n%s", productURL)
			err := sendTelegramMessage(msg)
			if err != nil {
				fmt.Println("❌ Error enviando mensaje de Telegram:", err)
			}
			break
		} else {
			fmt.Println("❌ No hay stock. Reintentando en", checkInterval)
		}

		time.Sleep(checkInterval)
	}
}

func checkStock() (bool, error) {
	req, err := http.NewRequest("GET", productURL, nil)
	if err != nil {
		return false, err
	}

	req.Header = http.Header{
		"User-Agent":      []string{userAgents[rand.Intn(len(userAgents))]},
		"Accept":          []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language": []string{"es-ES,es;q=0.9"},
		"Referer":         []string{"https://www.google.com/"},
		"Dnt":             []string{"1"},
	} // Evitar bloqueos básicos

	client := &http.Client{
		Timeout: 15 * time.Second, // Timeout más generoso
	}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	html := string(body)

	// Si contiene "No disponible", no hay stock
	if strings.Contains(html, "Agotado") {
		return false, nil
	}

	return true, nil
}

func sendTelegramMessage(text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramToken)

	payload := map[string]interface{}{
		"chat_id": telegramChatID,
		"text":    text,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("telegram error: %s", string(bodyBytes))
	}

	return nil
}

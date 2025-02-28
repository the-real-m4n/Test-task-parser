package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/geziyor/geziyor/export"
)

// Структура для хранения данных о товаре
type Product struct {
	ID              string  `json:"id"`
	ProductID       string  `json:"productId"`
	SKU             string  `json:"skuId"`
	Name            string  `json:"name"`
	Brand           string  `json:"brand"`
	Category        string  `json:"category"`
	Price           float64 `json:"price"`
	URL             string  `json:"url"`
	MeasurementUnit string  `json:"measurementUnit"`
}

func main() {
	cmd := exec.Command("C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe", "--remote-debugging-port=9222", "--disable-gpu", "--headless") // запуск отладочного режима chrome для работы программы
	err := cmd.Start()
	if err != nil {
		fmt.Println("ошибка при запуске Chrome")
	}
	geziyor.NewGeziyor(&geziyor.Options{
		BrowserEndpoint: "ws://localhost:9222", // порт
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Timeout:         30 * time.Second,
		LogDisabled:     false,

		StartRequestsFunc: func(g *geziyor.Geziyor) {
			g.GetRendered("https://www.okeydostavka.ru/msk/ovoshchi-i-frukty/ovoshchi", parseProducts)
		},
		Exporters: []export.Exporter{&export.JSON{}},
	}).Start()
	fmt.Println("Резульат сохранен в файл out.json в корневой папке")
}

func parseProducts(g *geziyor.Geziyor, r *client.Response) {
	r.HTMLDoc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptContent := s.Text()

		// Проверяем, содержит ли скрипт "var product = {"
		if strings.Contains(scriptContent, "var product = {") {
			re := regexp.MustCompile(`var product = ({.*?});`)
			match := re.FindStringSubmatch(scriptContent)

			if len(match) < 2 {
				log.Println("❌ JSON не найден в скрипте")
				return
			}

			// Приводим JSON к корректному виду
			jsonData := match[1]
			jsonData = fixJSON(jsonData) // Исправляем JSON

			log.Println("✅ Исправленный JSON:", jsonData)

			// Парсим JSON
			var product Product
			err := json.Unmarshal([]byte(jsonData), &product)
			if err != nil {
				log.Println("❌ Ошибка парсинга JSON:", err)
				return
			}

			// Записываем данные в экспорт
			g.Exports <- map[string]interface{}{
				"name":  product.Name,
				"price": product.Price,
				"link":  "https://www.okeydostavka.ru" + product.URL,
			}

			log.Printf("✅ Найден товар: %s - %.2f руб.", product.Name, product.Price)
		}
	})
}

// Функция исправления JSON
func fixJSON(input string) string {
	// 1. Добавляем кавычки к ключам
	reKeys := regexp.MustCompile(`(\w+):`)
	input = reKeys.ReplaceAllString(input, `"$1":`)

	// 2. Меняем одинарные кавычки на двойные
	input = strings.ReplaceAll(input, `'`, `"`)

	return input
}

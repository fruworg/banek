//
// анекдоты категории /b
//
// Руслан <im@fruw.org>, 2023
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	ID   int         `json:"id"`
	Text interface{} `json:"text"`
}

type Channel struct {
	Name     string    `json:"name"`
	ID       int       `json:"id"`
	Messages []Message `json:"messages"`
}

var (
	showVersion bool
	port        int
	content     string
	html        string
	jsonParse   string
	messages    Channel
	maxID       int
)

const (
	version = "1.0.0"
	cyan    = "\033[1;36m"
	reset   = "\033[0m"
)

func init() {
	flag.BoolVar(&showVersion, "v", false, "")
	flag.BoolVar(&showVersion, "version", false, "")
	flag.IntVar(&port, "p", 9999, "")
	flag.IntVar(&port, "port", 9999, "")
	flag.StringVar(&content, "c", "/etc/banek/content.json", "")
	flag.StringVar(&content, "content", "/etc/banek/content.json", "")
	flag.StringVar(&html, "h", "/etc/banek/template.html", "")
	flag.StringVar(&html, "html", "/etc/banek/template.html", "")
	flag.StringVar(&jsonParse, "j", "", "")
	flag.StringVar(&jsonParse, "json-parse", "", "")

	flag.Usage = func() {
		fmt.Printf("banek %s%s%s - анекдоты категории /b\n\n", cyan, version, reset)
		fmt.Printf("%s-h%s, %s--help%s       Показать справку banek и завершить работу\n", cyan, reset, cyan, reset)
		fmt.Printf("%s-v%s, %s--version%s    Показать версию banek и завершить работу\n", cyan, reset, cyan, reset)
		fmt.Printf("%s-p%s, %s--port%s       Порт, на котором будет запущен сервис (по-умолчанию 9999)\n", cyan, reset, cyan, reset)
		fmt.Printf("%s-c%s, %s--content%s    Путь до json-файла с анеками (по-умолчанию /etc/banek/content.json)\n", cyan, reset, cyan, reset)
		fmt.Printf("%s-h%s, %s--html%s       Путь до html-шаблона (по-умолчанию /etc/banek/template.html)\n", cyan, reset, cyan, reset)
		fmt.Printf("%s-j%s, %s--json-parse%s Спарсить выгрузку из группы в json\n\n", cyan, reset, cyan, reset)
		fmt.Println("Руслан, <im@fruw.org>, 2023")
	}
	flag.Parse()
}

func main() {
	if showVersion {
		fmt.Printf("banek %s%s%s\n", cyan, version, reset)
		return
	}

	if jsonParse != "" {
		parseJSON()
		return
	}

	rand.Seed(time.Now().UnixNano())

	err := loadMessages(content)
	if err != nil {
		fmt.Println("Ошибка чтения %d: %d", content, err)
		return
	}

	_, err = ioutil.ReadFile(html)
	if err != nil {
		fmt.Println("Ошибка чтения %d: %d", html, err)
		return
	}

	http.HandleFunc("/", handleRequest)
	if err != nil {
		fmt.Println("Ошибка сервиса:", err)
		return
	}

	fmt.Printf("Сервис запущен на порту :%d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}


func parseJSON() {
    jsonData, err := ioutil.ReadFile(jsonParse)
    if err != nil {
	fmt.Println("Ошибка чтения %d: %d", content, err)
        return
    }

    var channel Channel
    if err := json.Unmarshal(jsonData, &channel); err != nil {
        fmt.Println("Ошибка парсинга JSON:", err)
        return
    }

    messageNumber := 1
    for i, message := range channel.Messages {
        if message.Text != nil {
            switch text := message.Text.(type) {
            case string:
                if text != "" {
                    channel.Messages[i].ID = messageNumber
                    channel.Messages[i].Text = fmt.Sprintf("%s", text)
                    messageNumber++
                }
            case []interface{}:
                var newText []string
                for _, item := range text {
                    if str, ok := item.(string); ok && str != "" {
                        channel.Messages[i].ID = messageNumber
                        newText = append(newText, str)
                        messageNumber++
                    }
                }
                if len(newText) > 0 {
                    channel.Messages[i].Text = strings.Join(newText, " ")
                }
            }
        }
    }

    cleanedMessages := make([]Message, 0)
    for _, message := range channel.Messages {
        if message.Text != "" {
            cleanedMessages = append(cleanedMessages, message)
        }
    }
    channel.Messages = cleanedMessages

    formattedJSON, err := json.MarshalIndent(channel, "", "  ")
    if err != nil {
        fmt.Println("Ошибка форматирования JSON:", err)
        return
    }

    if err := ioutil.WriteFile(content, formattedJSON, 0644); err != nil {
        fmt.Printf("Ошибка записи JSON файла %d: %d", content, err)
        return
    }

    fmt.Println("JSON файл успешно записан по пути", content)
    return
}


func loadMessages(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &messages); err != nil {
		return err
	}

	maxID = determineMaxID()
	return nil
}

func determineMaxID() int {
	max := 0
	for _, message := range messages.Messages {
		if message.ID > max {
			max = message.ID
		}
	}
	return max
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[1:]

	id, err := strconv.Atoi(idStr)

	if err != nil || id < 1 || id > maxID {
		id = rand.Intn(maxID)
	}

	message := findMessageByID(id)

	userAgent := r.UserAgent()

	if strings.HasPrefix(userAgent, "curl") {
		if text, ok := message.Text.(string); ok {
			writePlainTextResponse(w, text, id)
		} else {
			fmt.Println("Ошибка: неверный Анек")
		}
	} else {
		prevID, nextID := getPrevNextIDs(id)
		htmlTemplate, _ := ioutil.ReadFile(html)
		if text, ok := message.Text.(string); ok {
			htmlContent := generateHTMLContent(w, htmlTemplate, id, text, prevID, nextID)
			writeHTMLResponse(w, htmlContent)
		} else {
			fmt.Println("Ошибка: неверный Анек")
		}
	}
}

func findMessageByID(id int) *Message {
	for _, message := range messages.Messages {
		if message.ID == id {
			return &message
		}
	}
	return nil
}

func getPrevNextIDs(id int) (int, int) {
	prevID := id - 1
	if id <= 1 {
		prevID = maxID
	}

	nextID := id + 1
	if id >= maxID {
		nextID = 1
	}
	return prevID, nextID
}

func generateHTMLContent(w http.ResponseWriter, htmlTemplate []byte, id int, text string, prevID int, nextID int) string {
	htmlContent := string(htmlTemplate)
	htmlContent = strings.ReplaceAll(htmlContent, "CURRENT_ANEK", strconv.Itoa(id))
	htmlContent = strings.ReplaceAll(htmlContent, "TEMPLATE_TEXT", text)
	htmlContent = strings.ReplaceAll(htmlContent, "PREV_ANEK", strconv.Itoa(prevID))
	htmlContent = strings.ReplaceAll(htmlContent, "NEXT_ANEK", strconv.Itoa(nextID))
	return htmlContent
}

func writePlainTextResponse(w http.ResponseWriter, text string, id int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Анек №" + strconv.Itoa(id) + "\n\n" + text + "\n"))
}

func writeHTMLResponse(w http.ResponseWriter, htmlContent string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlContent))
}

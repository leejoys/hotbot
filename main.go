package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	botToken := os.Getenv("hotToken")
	botApi := "https://api.telegram.org/bot"
	botUrl := botApi + botToken
	offset := 0
	for {
		updates, err := getUpdates(botUrl, offset)
		if err != nil {
			log.Println("main Error:", err)
		}
		for _, update := range updates {
			err := respond(botUrl, update)
			offset = update.UpdateId + 1
			if err != nil {
				log.Println("main respond Error:", err)
			}
		}
	}
}

func getUpdates(botUrl string, offset int) ([]Update, error) {
	resp, err := http.Get(botUrl + "/getUpdates" + "?offset=" + fmt.Sprint(offset))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var restResponse RestResponse
	err = json.Unmarshal(body, &restResponse)
	if err != nil {
		return nil, err
	}
	return restResponse.Result, nil
}

func respond(botUrl string, update Update) error {
	var botMessage BotMessage
	price, err := getPrice("https://ru.investing.com/currencies/" + update.Message.Text)
	if err != nil {
		return err
	}
	botMessage.ChatId = update.Message.Chat.ChatId
	botMessage.Text = price
	buf, err := json.Marshal(botMessage)
	if err != nil {
		return err
	}
	_, err = http.Post(botUrl+"/sendMessage", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	return nil
}

func getPrice(path string) (string, error) {
	res, err := http.Get(path)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}
	price := doc.Find(".instrument-price_instrument-price__3uw25")
	priceLast := price.Find(".text-2xl")
	return priceLast.Text(), nil
}

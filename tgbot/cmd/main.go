package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"tgbot/internal/env"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Tag struct {
	ID   uint	`json:"id"`
	Name string `json:"name"`
}

type Image struct {
	ID       uint
	Hash     string `json:"hash"`
	Filename string `json:"filename"`
	Tags     []string  `json:"tags"`
	Format   string `json:"format"`
	Width    uint   `json:"width"`
	Height   uint   `json:"height"`
}

var API_ORIGIN = env.GetEnvString("API_ORIGIN")
var API_LOCAL = env.GetEnvString("API_LOCAL")

const PAGE_SIZE = 30

func searchImages(query string, cursor int) ([]Image, error) {
	url := fmt.Sprintf("%s/images?cursor=%d&limit=%d&name=%s", API_LOCAL, cursor, PAGE_SIZE, url.QueryEscape(query))

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status %d", resp.StatusCode)
	}

	var res []Image
	err = json.NewDecoder(resp.Body).Decode(&res)
	return res, err
}

func handleInlineQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	query := update.InlineQuery.Query

	offset, err := 0, error(nil)

	if update.InlineQuery.Offset != "" {
		offset, err = strconv.Atoi(update.InlineQuery.Offset)
		if err != nil {
			offset = 0
		}
	}

	fmt.Printf("Query:%s, offset: %d\n", query, offset)
	images, err := searchImages(query, offset)
	if err != nil {
		println(err.Error())
		return
	}

	if len(images) == 0 {
		println("No images found")
		return
	}

	println(len(images), " images found")

	results := []tgbotapi.InlineQueryResultPhoto{}

	for _, img := range images {
		name := url.QueryEscape(img.Filename)
		photo := tgbotapi.NewInlineQueryResultPhotoWithThumb(
			fmt.Sprintf("%d", img.ID),
			API_ORIGIN+"/image/raw/"+name,
			API_ORIGIN+"/image/thumb/"+name+"."+img.Format,
		)
		photo.Title = img.Filename
		photo.Caption = img.Filename
		results = append(results, photo)
		// print(img.Filename, img.ID, " added\n")
	}

	nextOffset := ""
	if len(images) == PAGE_SIZE {
		nextOffset = fmt.Sprint(offset + PAGE_SIZE)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		Results:       make([]interface{}, len(results)),
		CacheTime:     100,
		NextOffset:    nextOffset,
		IsPersonal:    false,
	}

	for i, result := range results {
		inlineConf.Results[i] = result
	}

	resp, err := bot.Request(inlineConf)
	if err != nil {
		fmt.Println("Telegram request error:", err)
		return
	}

	if !resp.Ok {
		fmt.Println("Telegram API error:", resp.Description)
	}
}

var helloMessage string = "загружати картонкі: kartonko.lol\nшукати картонкі в тг @kartonkobot <запит>"

func handleMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, helloMessage)
	msg.ReplyToMessageID = update.Message.MessageID
	bot.Send(msg)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(env.GetEnvString("BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	fmt.Print("Bot is running...\n")
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleMessage(bot, update)
		}

		if update.InlineQuery != nil {
			handleInlineQuery(bot, update)
		}

	}
}

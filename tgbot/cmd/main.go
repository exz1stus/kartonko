package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"tgbot/internal/env"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Tag struct {
	Name string `json:"name"`
}

type Image struct {
	ID       uint
	Hash     string `json:"hash"`
	Filename string `json:"filename"`
	Tags     []Tag  `json:"tags"`
	Format   string `json:"format"`
	Width    uint   `json:"width"`
	Height   uint   `json:"height"`
}

type ApiResponse struct {
	Images []Image `json:"imageData"`
}

var API_ORIGIN = env.GetEnvString("API_ORIGIN")

const PAGE_SIZE = 10

func searchImages(query string, cursor int) ([]Image, error) {
	url := fmt.Sprintf("%s/images?cursor=%d&limit=%d&name=%s", API_ORIGIN, cursor, PAGE_SIZE, query)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res ApiResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	return res.Images, err
}

func handleInlineQuery(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	query := update.InlineQuery.Query

	offset := 0
	if update.InlineQuery.Offset != "" {
		offset, _ = strconv.Atoi(update.InlineQuery.Offset)
	}

	fmt.Printf("Query:%s, offset: %d\n", query, offset)
	images, err := searchImages(query, offset)
	if err != nil {
		print(err.Error())
		return
	}

	if len(images) == 0 {
		print("No images found\n")
		return
	}

	print(len(images), " images found\n")

	results := []tgbotapi.InlineQueryResultPhoto{}

	for _, img := range images {
		photo := tgbotapi.NewInlineQueryResultPhotoWithThumb(
			fmt.Sprint(img.ID),
			API_ORIGIN+"/raw-image/"+img.Filename,       // photo_url
			API_ORIGIN+"/raw-image/thumb/"+img.Filename, // thumb_url
		)
		photo.Title = img.Filename
		photo.Caption = img.Filename
		results = append(results, photo)
	}

	nextOffset := ""
	if len(images) == PAGE_SIZE {
		nextOffset = fmt.Sprint(offset + PAGE_SIZE)
	}

	inlineConf := tgbotapi.InlineConfig{
		InlineQueryID: update.InlineQuery.ID,
		Results:       make([]interface{}, len(results)),
		CacheTime:     10,
		NextOffset:    nextOffset,
		IsPersonal:    true,
	}

	for i, result := range results {
		inlineConf.Results[i] = result
	}
	bot.Request(inlineConf)
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
		if update.InlineQuery == nil {
			continue
		}

		handleInlineQuery(bot, update)
	}
}

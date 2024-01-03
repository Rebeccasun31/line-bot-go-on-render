// Copyright 2016 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"encoding/csv"
	"strings"
	"math/rand"
	"time"

	"github.com/line/line-bot-sdk-go/v8/linebot"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

type Drink struct {
	Name string
	Sweet string
	Ice string
	Price string
}

// var list_len int = 1

var drinklist = []Drink {
	Drink {
		Name: "白開水",
		Sweet: "無糖",
		Ice: "去冰",
		Price: "0",
	},
}

func read_csv() {
	file, err := os.Open("code/drink.csv")
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		log.Println("Error:", err)
		return
	}

	for idx, record := range records {
		if idx == 0 {
			continue
		}

		// list_len += 1

		drink := Drink{
			Name: record[0] + record[1],
			Sweet: "微糖",
			Ice: "微冰",
			Price: record[2],
		}

		drinklist = append(drinklist, drink)
	}

}

func addDrink(input string) string {
	str := strings.Fields(input)
	if len(str) != 5 {
		return "輸入格式錯誤！"
	}
	newDrink := Drink {
		Name: str[1],
		Sweet: str[2],
		Ice: str[3],
		Price: str[4],
	}

	for _, drink := range drinklist {
		if drink.Name == newDrink.Name {
			return "該飲料已在清單中！"
		}
	}
	drinklist = append(drinklist, newDrink)
	return "新增成功！"
}

func delDrink(input string) string {
	str := strings.Fields(input)
	if len(str) != 2 {
		return "輸入格式錯誤！"
	}
	var target string = str[1]

	var foundIndex = -1

	for i, drink := range drinklist {
		if target == drink.Name {
			foundIndex = i
			break
		}
	}
	if foundIndex == -1 {
		return "找不到！"
	}

	drinklist = append(drinklist[:foundIndex], drinklist[foundIndex+1:]...)

	return "刪除成功！"
}

func search(input string) string {
	str := strings.Fields(input)
	if len(str) != 2 {
		return "輸入格式錯誤！"
	}
	var target string = str[1]

	var foundIndex = -1

	for i, drink := range drinklist {
		if target == drink.Name {
			foundIndex = i
			break
		}
	}
	if foundIndex == -1 {
		return "找不到！"
	}

	reply := fmt.Sprintf(
		"%s %s %s，價格：%s元", drinklist[foundIndex].Name, drinklist[foundIndex].Sweet, drinklist[foundIndex].Ice, drinklist[foundIndex].Price)
	return reply
}

func main() {
	read_csv()
	channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
	bot, err := messaging_api.NewMessagingApiAPI(
		os.Getenv("LINE_CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		log.Println("/callback called...")

		cb, err := webhook.ParseRequest(channelSecret, req)
		if err != nil {
			log.Printf("Cannot parse request: %+v\n", err)
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}

		log.Println("Handling events...")
		for _, event := range cb.Events {
			log.Printf("/callback called%+v...\n", event)

			switch e := event.(type) {
			case webhook.MessageEvent:
				switch message := e.Message.(type) {
				// 收到的是文字訊息
				case webhook.TextMessageContent:
					var reply string
					if message.Text[0] == '1' {
						reply = addDrink(message.Text)

					} else if message.Text[0] == '2' {
						reply = delDrink(message.Text)
						
					} else if message.Text[0] == '3' {
						reply = search(message.Text)

					} else if message.Text[0] == '4' { //latest 10
						reply = ""
						list_len := len(drinklist) - 1

						for i := 0; i < 10 && i < list_len; i++ {
							cur := fmt.Sprintf(
								"%s %s %s，價格：%s元\n", drinklist[list_len-i].Name, drinklist[list_len-i].Sweet, drinklist[list_len-i].Ice, drinklist[list_len-i].Price)

							reply += cur
						}

						if len(reply) > 0 {
							reply = reply[:len(reply)-1]
						}

					} else {
						rand.Seed(time.Now().UnixNano())
						list_len := len(drinklist) - 1
						idx := rand.Intn(list_len)
						reply = fmt.Sprintf(
							"推薦飲料：%s %s %s，價格：%s元", drinklist[idx].Name, drinklist[idx].Sweet, drinklist[idx].Ice, drinklist[idx].Price)
					}

					// 回覆
					if _, err = bot.ReplyMessage(
						&messaging_api.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages: []messaging_api.MessageInterface{
								messaging_api.TextMessage{
									Text: reply,
								},
							},
						},
					); err != nil {
						log.Print(err)
					} else {
						log.Println("Sent text reply.")
					}

				// 收到的是貼圖
				case webhook.StickerMessageContent:
					var reply string

					rand.Seed(time.Now().UnixNano())
					list_len := len(drinklist) - 1
					idx := rand.Intn(list_len)
					reply = fmt.Sprintf(
						"推薦飲料：%s %s %s，價格：%s元", drinklist[idx].Name, drinklist[idx].Sweet, drinklist[idx].Ice, drinklist[idx].Price)

					if _, err = bot.ReplyMessage(
						&messaging_api.ReplyMessageRequest{
							ReplyToken: e.ReplyToken,
							Messages: []messaging_api.MessageInterface{
								messaging_api.TextMessage{
									Text: reply,
								},
							},
						}); err != nil {
						log.Print(err)
					} else {
						log.Println("Sent sticker reply.")
					}
				default:
					log.Printf("Unsupported message content: %T\n", e.Message)
				}
			default:
				log.Printf("Unsupported message: %T\n", event)
			}
		}
	})

	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	fmt.Println("http://localhost:" + port + "/")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
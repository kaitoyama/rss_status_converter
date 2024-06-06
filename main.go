package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/sashabaranov/go-openai"
	traq "github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	payload "github.com/traPtitech/traq-ws-bot/payload"
	"google.golang.org/api/option"
)

func SubGenerateText(input string) string {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleUser,
					Content: `絵文字を交えながら以下のテンプレートに従って入力を日本のエンジニア向けに変換してください。
                                        テンプレート
                                        ## 🛠️ {サービス名 GitHub や OpenAI や DeepL など}:{[インシデントタイトル](url)}
                                        タイムライン
                                        🗓️ {日付} {時刻} {タイムゾーン JTCに変換すること}
                                        {状況 ✅ 解決済み や 🔄 更新 や 🔍 調査中} - {具体的な状況の説明}
                                        
                                        入力` + input,
				},
			},
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	return (resp.Choices[0].Message.Content)
}

func GenerateText(input string) string {
	log.Println("GenerateText function called")
	ctx := context.Background()

	GeminiAPIKey := os.Getenv("GOOGLE_ACCESS_TOKEN")
	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey(GeminiAPIKey))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// For text-only input, use the gemini-pro model
	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := genai.Text(`絵文字を交えながら以下のテンプレートに従って入力を日本のエンジニア向けに変換してください。
テンプレート
## 🛠️ {サービス名 GitHub や OpenAI や DeepL など}:{[インシデントタイトル](url)}
タイムライン
🗓️ {日付} {時刻} {タイムゾーン JTCに変換すること}
{状況 ✅ 解決済み や 🔄 更新 や 🔍 調査中} - {具体的な状況の説明}

入力` + input)
	resp, err := model.GenerateContent(ctx, prompt)
	if err != nil {
		log.Println(err)
		return SubGenerateText(input)
	}
	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		result += fmt.Sprint(part)
	}
	return result
}

func main() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Println("Error loading .env file")
	// }
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: os.Getenv("ACCESS_TOKEN"),
	})
	if err != nil {
		panic(err)
	}

	bot.OnMessageCreated(func(p *payload.MessageCreated) {
		log.Println("Received MESSAGE_CREATED event: " + p.Message.Text)
		if p.Message.User.Name == "kaitoyama" || p.Message.User.Name == "BOT_yuri" {
			message := GenerateText(p.Message.Text)
			_, _, err := bot.API().
				MessageApi.
				PostMessage(context.Background(), p.Message.ChannelID).
				PostMessageRequest(traq.PostMessageRequest{
					Content: message,
				}).
				Execute()
			if err != nil {
				log.Println(err)
			}
		}
	})

	if err := bot.Start(); err != nil {
		panic(err)
	}
}

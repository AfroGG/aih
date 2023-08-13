package main

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	//"github.com/go-rod/rod/lib/utils"
	//"github.com/go-rod/stealth"

	//"bufio"
	//"bytes"
	"context"
	//"encoding/json"
	"fmt"
	//"github.com/Databingo/EdgeGPT-Go"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	//"github.com/google/uuid"
	"github.com/manifoldco/promptui"
	"github.com/peterh/liner"
	"github.com/rivo/tview"
	openai "github.com/sashabaranov/go-openai"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	//"io"
	"io/ioutil"
	//"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var color_bard = tcell.ColorDarkCyan
var color_bing = tcell.ColorDarkMagenta
var color_chat = tcell.ColorWhite
var color_chatapi = tcell.ColorWhite
var color_claude = tcell.ColorYellow
var color_huggingchat = tcell.ColorDarkMagenta

func clear() {
	switch runtime.GOOS {
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func multiln_input(Liner *liner.State, prompt string) string {
	// For recognize multipile lines input module
	// |--------------------------|------
	// |recording && input        | action
	// |--------------------------|------
	// |false && == "" or x       | record; break;
	// |false && != "<<"          | record; break;
	// |false && == "<<" + ">>"   | record; break; rm << >>;
	// |false && == "<<"          | record; true; rm <<;
	// |true  && == "" or x       | record;
	// |true  && != ">>"          | record;
	// |true  && == ">>"          | record; break; rm >>;
	// |--------------------------|------

	var ln string
	var lns []string
	recording := false
	for {
		if ln == "" && !recording {
			ln, _ = Liner.Prompt(prompt)
		} else {
			ln, _ = Liner.Prompt("")
		}
		ln = strings.Trim(ln, " ")
		if !recording && (ln == "" || len(ln) == 1) {
			lns = append(lns, ln)
			break
		} else if !recording && ln[:2] != "<<" {
			lns = append(lns, ln)
			break
		} else if !recording && ln[:2] == "<<" && len(ln) >= 4 && ln[len(ln)-2:] == ">>" {
			lns = append(lns, ln[2:len(ln)-2])
			break
		} else if !recording && ln[:2] == "<<" {
			recording = true
			lns = append(lns, ln[2:])
		} else if recording && (ln == "" || len(ln) == 1) {
			lns = append(lns, ln)
		} else if recording == true && ln[len(ln)-2:] != ">>" {
			lns = append(lns, ln)
		} else if recording == true && ln[len(ln)-2:] == ">>" {
			recording = false
			lns = append(lns, ln[:len(ln)-2])
			break
		}
	}

	long_str := strings.Join(lns, "\n")
	return long_str
}

// Write response RESP to clipboard
func save2clip_board(rs string) {
	err := clipboard.WriteAll(rs)
	if err != nil {
		panic(err)
	}
}

func main() {
	// Create prompt for user input
	Liner := liner.NewLiner()
	defer Liner.Close()

	// Use RESP for record response per time
	var RESP string

	// Read Aih Configure
	aih_json, err := ioutil.ReadFile("aih.json")
	if err != nil {
		err = ioutil.WriteFile("aih.json", []byte(""), 0644)
	}

	// Read Proxy
	Proxy := gjson.Get(string(aih_json), "proxy").String()
	fmt.Println("Proxy:", Proxy)
	Proxy = "socks5://127.0.0.1:7890"

	// Set proxy for system of current program
	os.Setenv("http_proxy", Proxy)
	os.Setenv("https_proxy", Proxy)

	// Set proxy for rod
	//proxy_url := launcher.New().Proxy(Proxy).Delete("use-mock-keychain").MustLaunch()
	//proxy_url := launcher.New().
	//	StartURL("about:blank").
	//	Proxy(Proxy).
	//	MustLaunch()
	//	//UserDataDir("data").

	proxy_url := launcher.NewUserMode().
		Proxy(Proxy).
		//Leakless(true).// indepent tab | work with UserDataDir()
		//UserDataDir("data").// indepent tab + data
		//Set("disable-default-apps").
		//Headless(true).
		MustLaunch()

	// Open rod browser
	var browser *rod.Browser
	if Proxy != "" {
		browser = rod.New().
			Trace(true).
			ControlURL(proxy_url).
			Timeout(60 * 24 * time.Minute).
			MustConnect()
		//.NoDefaultDevice()
	} else {
		browser = rod.New().
			Trace(true).
			Timeout(60 * 24 * time.Minute).
			MustConnect()
		//.NoDefaultDevice()
	}

	// Read user.json
	user_json, _ := ioutil.ReadFile("user.json")
	if err != nil {
		err = ioutil.WriteFile("user.json", []byte(""), 0644)
	}

	// Read user/password
	var chatgpt_user string
	var chatgpt_password string
	var bard_user string
	var bard_password string
	chatgpt_user = gjson.Get(string(user_json), "chatgpt.user").String()
	chatgpt_password = gjson.Get(string(user_json), "chatgpt.password").String()
	//chatgpt_user = ""
	//chatgpt_password = ""
	bard_user = gjson.Get(string(user_json), "bard.user").String()
	bard_password = gjson.Get(string(user_json), "bard.password").String()

	fmt.Println(chatgpt_user)
	fmt.Println(chatgpt_password)
	fmt.Println(bard_user)
	fmt.Println(bard_password)

	// Test Proxy
	//TEST_PROXY:
	//	fmt.Println("Checking network accessing...")
	//	ops1 := googlesearch.SearchOptions{Limit: 12}
	//	_, err = googlesearch.Search(nil, "BTC", ops1)
	//	if err != nil {
	//		fmt.Println("Need proxy to access GoogleBard, BingChat, ChatGPT")
	//		proxy, _ := Liner.Prompt("Please input proxy: ")
	//		if proxy == "" {
	//			goto TEST_PROXY
	//		}
	//		aihj, err := ioutil.ReadFile("aih.json")
	//		new_aihj, _ := sjson.Set(string(aihj), "proxy", proxy)
	//		err = ioutil.WriteFile("aih.json", []byte(new_aihj), 0644)
	//		if err != nil {
	//			fmt.Println("Save failed.")
	//		}
	//		fmt.Println("Please restart Aih for using proxy...")
	//		Liner.Close()
	//		syscall.Exit(0)
	//
	//	}

	// Set up client of OpenAI API
	key := gjson.Get(string(aih_json), "key")
	OpenAI_Key := key.String()
	config := openai.DefaultConfig(OpenAI_Key)
	client := openai.NewClientWithConfig(config)
	messages := make([]openai.ChatCompletionMessage, 0)

	//////////////////////0////////////////////////////
	// Set up client of ChatGPT (chromedriver version)

	//////////////////////1////////////////////////////
	// Set up client of Bard (chromedriver version)
	var page_bard *rod.Page
	var relogin_bard bool
	channel_bard := make(chan string)
	go func() {
		page_bard = browser.MustPage("https://bard.google.com")
		for {
			if page_bard.Timeout(10 * time.Second).MustHasX("//textarea[@id='mat-input-0']") {
				relogin_bard = false
				break
			}
			if page_bard.Timeout(10 * time.Second).MustHasX("//span[contains(text(), 'Sign in')]") {
				relogin_bard = true
				break
			}
			time.Sleep(time.Second)
		}
		if relogin_bard == true {
			fmt.Println("✘ Bard")
		}
		if relogin_bard == false {
			fmt.Println("✔ Bard")
			for {
				select {
				case question := <-channel_bard:
					//fmt.Println("question:", question)
					page_bard.MustElementX("//textarea[@id='mat-input-0']").MustWaitVisible().MustInput(question)
					page_bard.MustElementX("//button[@mattooltip='Submit']").MustClick()
					page_bard.MustElementX("//img[contains(@src, 'https://www.gstatic.com/lamda/images/sparkle_thinking_v2_e272afd4f8d4bbd25efe.gif')]").MustWaitVisible()
					img := page_bard.MustElementX("//img[contains(@src, 'https://www.gstatic.com/lamda/images/sparkle_resting_v2_1ff6f6a71f2d298b1a31.gif')]").MustWaitVisible()
					response := img.MustElementX("ancestor::model-response").MustWaitVisible()
					answer := response.MustText()
					answer = strings.Replace(answer, "View other drafts", "", -1)
					answer = strings.Replace(answer, "Regenerate draft", "", -1)
					answer = strings.Replace(answer, "thumb_up", "", -1)
					answer = strings.Replace(answer, "thumb_down", "", -1)
					answer = strings.Replace(answer, "volume_up", "", -1)
					answer = strings.Replace(answer, "more_vert", "", -1)
					answer = strings.Replace(answer, "Google it", "", -1)
					answer = answer[8 : len(answer)-12]
					channel_bard <- answer
				}
			}
		}

	}()
	//////////////////////2////////////////////////////
	// Set up client of Claude2 (Rod version)
	var page_claude *rod.Page
	var relogin_claude bool
	channel_claude := make(chan string)
	go func() {
		page_claude = browser.MustPage("https://claude.ai")
		for {
			if page_claude.Timeout(10 * time.Second).MustHasX("//p[@data-placeholder='Message Claude or search past chats...']") {
				page_claude.MustElementX("//div[contains(text(), 'Start a new chat')]").MustClick()
				page_claude.MustElementX("//p[@data-placeholder='Message Claude...']").MustWaitVisible()
				relogin_claude = false
				break
			}
			if page_claude.Timeout(10 * time.Second).MustHasX("//h2[contains(text(), 'Talk to Claude')]") {
				relogin_claude = true
				break
			}
			time.Sleep(time.Second)
		}
		if relogin_claude == true {
			fmt.Println("✘ Claude")
		}
		if relogin_claude == false {
			fmt.Println("✔ Claude")
			for {
				select {
				case question := <-channel_claude:
					//fmt.Println("question:", question)
					page_claude.MustElementX("//p[@data-placeholder='Message Claude...']").MustInput(question)
					page_claude.MustElementX("//button[@aria-label='Send Message']").MustClick()
					retry_icon := page_claude.MustElement("svg path[d='M224,128a96,96,0,0,1-94.71,96H128A95.38,95.38,0,0,1,62.1,197.8a8,8,0,0,1,11-11.63A80,80,0,1,0,71.43,71.39a3.07,3.07,0,0,1-.26.25L44.59,96H72a8,8,0,0,1,0,16H24a8,8,0,0,1-8-8V56a8,8,0,0,1,16,0V85.8L60.25,60A96,96,0,0,1,224,128Z']").MustWaitVisible()
					content := retry_icon.MustElementX("preceding::div[2]")
					answer := content.MustText()
					channel_claude <- answer
				}
			}
		}

	}()

	//////////////////////3////////////////////////////
	// Set up client of huggingchat (chromedriver version)

	//////////////////////4////////////////////////////
	// Set up client of chatgpt (rod version)
	var page_chatgpt *rod.Page
	var relogin_chatgpt bool
	channel_chatgpt := make(chan string)
	go func() {
		//page_chatgpt = stealth.MustPage(browser)
		page_chatgpt = browser.MustPage("https://chat.openai.com")
		for {
			if page_chatgpt.Timeout(10 * time.Second).MustHasX("//textarea[@id='prompt-textarea']") {
				relogin_chatgpt = false
				break
			}
			if page_chatgpt.Timeout(10 * time.Second).MustHasX("//div[contains(text(), 'Log in with your OpenAI account to continue')]") {
				relogin_chatgpt = true
				break
			}
			time.Sleep(time.Second)
		}

		if relogin_chatgpt == true {
			fmt.Println("✘ ChatGPT")
			//page_chatgpt.MustElementX("//div[contains(text(), 'Welcome to ChatGPT')] | //h2[contains(text(), 'Get started')]").MustWaitVisible()
			//page_chatgpt.MustElementX("//div[not(contains(@class, 'mb-4')) and contains(text(), 'Log in')]").MustClick()
			//utils.Sleep(1.5)
			//page_chatgpt.MustElementX("//input[@id='username']").MustWaitVisible().MustInput(chatgpt_user)
			//utils.Sleep(1.5)
			//page_chatgpt.MustElementX("//button[contains(text(), 'Continue')]").MustClick()
			//utils.Sleep(1.5)
			//page_chatgpt.MustElementX("//input[@id='password']").MustWaitVisible().MustInput(chatgpt_password)
			//utils.Sleep(1.5)
			//page_chatgpt.MustElementX("//button[not(contains(@aria-hidden, 'true')) and contains(text(), 'Continue')]").MustClick()
			////page_chatgpt.MustElementX("//h4[contains(text(), 'This is a free research preview.')]").MustWaitVisible()
			////utils.Sleep(1.5)
			////page_chatgpt.MustElementX("//button/div[contains(text(), 'Next')]").MustClick()
			////page_chatgpt.MustElementX("//h4[contains(text(), 'How we collect data')]").MustWaitVisible()
			////utils.Sleep(1.5)
			////page_chatgpt.MustElementX("//button/div[contains(text(), 'Next')]").MustClick()
			////page_chatgpt.MustElementX("//h4[contains(text(), 'love your feedback!')]").MustWaitVisible()
			////utils.Sleep(1.5)
			////page_chatgpt.MustElementX("//button/div[contains(text(), 'Done')]").MustClick()
			////utils.Sleep(1.5)
			//page_chatgpt.MustElementX("//a[contains(text(), 'New chat')]").MustWaitVisible().MustClick()
			//page_chatgpt.MustElementX("//textarea[@id='prompt-textarea']").MustWaitVisible()
			//utils.Sleep(1.5)
			//page_chatgpt.MustElementX("//textarea[@id='prompt-textarea']").MustInput("hello")
			//utils.Sleep(1.5)
			//sends := page_chatgpt.MustElements("button:last-of-type svg path[d='M.5 1.163A1 1 0 0 1 1.97.28l12.868 6.837a1 1 0 0 1 0 1.766L1.969 15.72A1 1 0 0 1 .5 14.836V10.33a1 1 0 0 1 .816-.983L8.5 8 1.316 6.653A1 1 0 0 1 .5 5.67V1.163Z']")
			//sends[len(sends)-1].MustClick()
			//page_chatgpt.MustElement("svg:last-of-type path[d='M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15']").MustWaitVisible()
			//fmt.Println("Retry icon show")
			//page_chatgpt.MustElementX("(//div[contains(@class, 'group w-full')])[last()]").MustText()
			//fmt.Println("✔ ChatGPT Ready")
		}
		if relogin_chatgpt == false {
			fmt.Println("✔ ChatGPT")
			for {
				select {
				case question := <-channel_chatgpt:
					//fmt.Println("question:", question)
					page_chatgpt.MustElementX("//textarea[@id='prompt-textarea']").MustWaitVisible().MustInput(question)
					sends := page_chatgpt.MustElements("button:last-of-type svg path[d='M.5 1.163A1 1 0 0 1 1.97.28l12.868 6.837a1 1 0 0 1 0 1.766L1.969 15.72A1 1 0 0 1 .5 14.836V10.33a1 1 0 0 1 .816-.983L8.5 8 1.316 6.653A1 1 0 0 1 .5 5.67V1.163Z']")
					sends[len(sends)-1].MustClick()
					page_chatgpt.MustElement("svg:last-of-type path[d='M20.49 9A9 9 0 0 0 5.64 5.64L1 10m22 4l-4.64 4.36A9 9 0 0 1 3.51 15']").MustWaitVisible()
					//fmt.Println("Retry icon show")
					answer := page_chatgpt.MustElementX("(//div[contains(@class, 'group w-full')])[last()]").MustText()
					channel_chatgpt <- answer
				}
			}
		}
	}()

	// Clean screen
	clear()

	// Welcome to Aih
	welcome := `
╭ ────────────────────────────── ╮
│    Welcome to Aih              │ 
│    Type .help for help         │ 
╰ ────────────────────────────── ╯ `
	fmt.Println(welcome)

	max_tokens := 4097
	used_tokens := 0
	left_tokens := 0
	speak := 0
	role := ".bard"
	uInput := ""
	//price := ""
	chat_mode := openai.GPT3Dot5Turbo
	chat_completion := true

	// Start loop to read user input
	for {
		// Re-read user input history
		if f, err := os.Open(".history"); err == nil {
			Liner.ReadHistory(f)
			f.Close()
		}

		prompt := strconv.Itoa(left_tokens) + role + "> "
		userInput := multiln_input(Liner, prompt)
		//fmt.Println("userInput:", userInput)

		// Check Aih commands
		switch userInput {
		case "":
			continue
		case ".proxy":
			proxy, _ := Liner.Prompt("Please input your proxy:")
			if proxy == "" {
				continue
			}
			aihj, err := ioutil.ReadFile("aih.json")
			new_aihj, _ := sjson.Set(string(aihj), "proxy", proxy)
			err = ioutil.WriteFile("aih.json", []byte(new_aihj), 0644)
			if err != nil {
				fmt.Println("Save failed.")
			}
			fmt.Println("Please restart Aih for using proxy")
			Liner.Close()
			syscall.Exit(0)
		case ".help":
			fmt.Println("                           ")
			fmt.Println(" .               Select AI mode of Bard/Bing/ChatGPT/Claude")
			fmt.Println(" .key            Set cookie of Bard/Bing/ChatGPT/Claude")
			fmt.Println(" .proxy          Set proxy")
			fmt.Println(" <<              Start multiple lines input")
			fmt.Println(" >>              End multiple lines input")
			fmt.Println(" ↑               Previous input")
			fmt.Println(" ↓               Next input")
			fmt.Println(" .c or .clear    Clear screen")
			fmt.Println(" .h or .history  Show history")
			fmt.Println(" j               Scroll down")
			fmt.Println(" k               Scroll up")
			fmt.Println(" gg              Scroll top")
			fmt.Println(" G               Scroll bottom")
			fmt.Println(" q or Enter      Back to conversation")
			fmt.Println(" .help           Help")
			fmt.Println(" .exit           Exit")
			fmt.Println(" .new            New conversation of ChatGPT")
			fmt.Println(" .speak          Voice speak context (MasOS only)")
			fmt.Println(" .quiet          Not speak")
			fmt.Println("                           ")
			fmt.Println("                           ")
			continue
		case ".c", ".clear":
			clear()
			continue
		case ".h", ".history":
			cnt, _ := ioutil.ReadFile("history.txt")
			printer(color_chat, string(cnt), true)
			continue
		case ".exit":
			switch runtime.GOOS {
			case "linux", "darwin":
				cmd := exec.Command("pkill", "-f", "undetected_chromedriver")
				err = cmd.Run()
				if err != nil {
					fmt.Println(err)
				}
			case "windows":
				cmd := exec.Command("taskkill", "/IM", "undetected_chromedriver", "/F")
				err = cmd.Run()
				if err != nil {
					fmt.Println(err)
				}
			}

			return
		case ".new":
			// For role .chat
			//conversation_id = ""
			//parent_id = ""
			// For role .chatapi
			messages = make([]openai.ChatCompletionMessage, 0)
			//max_tokens = 4097
			used_tokens = 0
			left_tokens = max_tokens - used_tokens
			continue
		case ".", "/":
			proms := promptui.Select{
				Label: "Select AI mode to chat",
				Size:  10,
				Items: []string{
					"Bard",
					//"Bing",
					"ChatGPT",
					"Claude",
					"HuggingChat",
					"ChatGPT API gpt-3.5-turbo, $0.002/1K tokens",
					"ChatGPT API gpt-4 8K Prompt, $0.03/1K tokens",
					"ChatGPT API gpt-4 8K Completion, $0.06/1K tokens",
					"ChatGPT API gpt-4 32K Prompt, $0.06/1K tokens",
					"ChatGPT API gpt-4 32K Completion, $0.12/1K tokens",
					"Exit",
				},
			}

			_, ai, err := proms.Run()
			if err != nil {
				panic(err)
			}

			switch ai {
			case "Bard":
				role = ".bard"
				left_tokens = 0
				continue
			case "Bing":
				role = ".bing"
				left_tokens = 0
				continue
			case "ChatGPT":
				role = ".chat"
				left_tokens = 0
				continue
			case "Claude":
				role = ".claude"
				left_tokens = 0
				continue
			case "HuggingChat":
				role = ".huggingchat"
				left_tokens = 0
				continue
			case "ChatGPT API gpt-3.5-turbo, $0.002/1K tokens":
				role = ".chatapi"
				chat_mode = openai.GPT3Dot5Turbo
				max_tokens = 4097
				used_tokens = 0
				left_tokens = max_tokens - used_tokens
				chat_completion = true
				continue
			case "ChatGPT API gpt-4 8K Prompt, $0.03/1K tokens":
				role = ".chatapi"
				chat_mode = openai.GPT4
				max_tokens = 8192
				used_tokens = 0
				left_tokens = max_tokens - used_tokens
				chat_completion = false
				continue
			case "ChatGPT API gpt-4 8K Completion, $0.06/1K tokens":
				role = ".chatapi"
				chat_mode = openai.GPT4
				max_tokens = 8192
				used_tokens = 0
				left_tokens = max_tokens - used_tokens
				chat_completion = true
				continue
			case "ChatGPT API gpt-4 32K Prompt, $0.06/1K tokens":
				role = ".chatapi"
				chat_mode = openai.GPT432K
				max_tokens = 32768
				used_tokens = 0
				left_tokens = max_tokens - used_tokens
				chat_completion = false
				continue
			case "ChatGPT API gpt-4 32K Completion, $0.12/1K tokens":
				role = ".chatapi"
				chat_mode = openai.GPT432K
				max_tokens = 32768
				used_tokens = 0
				left_tokens = max_tokens - used_tokens
				chat_completion = true
				continue
			case "Exit":
				continue
			}
		case ".key":
			prom := promptui.Select{
				Label: "Select:",
				Size:  6,
				Items: []string{
					"Set Bard Cookie",
					"Set ChatGPT Cookie",
					"Set Claude Cookie",
					"Set HuggingChat Cookie",
					"Set ChatGPT API Key",
					"Exit",
				},
			}

			_, keyy, err := prom.Run()
			if err != nil {
				panic(err)
			}

			switch keyy {
			case "Set Bard Cookie":
				role = ".bard"
				goto BARD
			case "Set ChatGPT Cookie":
				role = ".chat"
				goto CHAT
			case "Set ChatGPT API Key":
				OpenAI_Key = ""
				role = ".chatapi"
				goto CHATAPI
			case "Set Claude Cookie":
				role = ".claude"
				goto CLAUDE
			case "Set HuggingChat Cookie":
				role = ".huggingchat"
				goto HUGGINGCHAT
			case "Exit":
				continue
			}

		case ".speak":
			speak = 1
			continue
		case ".quiet":
			speak = 0
			continue
		default:
			// Re-read user input history in case other process alternated
			if f, err := os.Open(".history"); err == nil {
				Liner.ReadHistory(f)
				f.Close()
			}
			// Record user input without Aih commands
			uInput = strings.Replace(userInput, "\r", "\n", -1)
			uInput = strings.Replace(uInput, "\n", " ", -1)
			Liner.AppendHistory(uInput)
			// Persistent user input
			if f, err := os.Create(".history"); err == nil {
				Liner.WriteHistory(f)
				f.Close()
			}

		}

	BARD:
		// Check role for correct actions
		if role == ".bard" {
			//fmt.Println("type question:", userInput)
			if relogin_bard == true {
				fmt.Println("Login Bard please.")
			} else {
				page_bard.Activate()
				channel_bard <- userInput
				answer := <-channel_bard

				// Print the response to the terminal
				RESP = strings.TrimSpace(answer)
				//used_tokens = resp.Usage.TotalTokens
				//left_tokens = max_tokens - used_tokens
				//printer_chat.Println(RESP)
				printer(color_bard, RESP, false)
			}

		}

	CLAUDE:
		// Check role for correct actions
		if role == ".claude" {
			if relogin_claude == true {
				fmt.Println("Login Claude please.")
			} else {
				page_claude.Activate()
				channel_claude <- userInput
				answer := <-channel_claude

				// Print the response to the terminal
				RESP = strings.TrimSpace(answer)
				//used_tokens = resp.Usage.TotalTokens
				//left_tokens = max_tokens - used_tokens
				//printer_chat.Println(RESP)
				printer(color_claude, RESP, false)
			}

		}
	CHAT:
		if role == ".chat" {
			//fmt.Println("type question:", userInput)
			if relogin_chatgpt == true {
				fmt.Println("Login ChatGPT please.")
			} else {
				page_chatgpt.Activate()
				channel_chatgpt <- userInput
				answer := <-channel_chatgpt

				// Print the response to the terminal
				RESP = strings.TrimSpace(answer)
				//used_tokens = resp.Usage.TotalTokens
				//left_tokens = max_tokens - used_tokens
				//printer_chat.Println(RESP)
				printer(color_chatapi, RESP, false)
			}

		}

	HUGGINGCHAT:
		if role == ".huggingchat" {

		}
	CHATAPI:
		if role == ".chatapi" {
			// Check ChatGPT API Key
			if OpenAI_Key == "" {
				OpenAI_Key, _ = Liner.Prompt("Please input your OpenAI Key: ")
				if OpenAI_Key == "" {
					continue
				}
				aihj, err := ioutil.ReadFile("aih.json")
				new_aihj, _ := sjson.Set(string(aihj), "key", OpenAI_Key)
				err = ioutil.WriteFile("aih.json", []byte(new_aihj), 0644)
				if err != nil {
					fmt.Println("Save failed.")
				}
				// Renew ChatGPT client with key
				config = openai.DefaultConfig(OpenAI_Key)
				client = openai.NewClientWithConfig(config)
				messages = make([]openai.ChatCompletionMessage, 0)
				left_tokens = 0
				continue
			}
			// Porcess input
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: userInput,
			})

			// Generate a response from ChatGPT
			resp, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model:    chat_mode, //openai.GPT3Dot5Turbo,
					Messages: messages,
				},
			)

			if err != nil {
				fmt.Println(">>>", err)
				continue
			}

			// Record in coversation context
			if chat_completion {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: RESP,
				})
			}

			// Print the response to the terminal
			RESP = strings.TrimSpace(resp.Choices[0].Message.Content)
			used_tokens = resp.Usage.TotalTokens
			left_tokens = max_tokens - used_tokens
			//printer_chat.Println(RESP)
			printer(color_chatapi, RESP, false)

		}

		// -------------for all AI's RESP---------------

		// Persistent conversation uInput + response
		if fs, err := os.OpenFile("history.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666); err == nil {
			time_string := time.Now().Format("2006-01-02 15:04:05")
			_, err = fs.WriteString("--------------------\n")
			_, err = fs.WriteString(time_string + role + "\n\nQuestion:\n" + uInput + "\n\n")
			_, err = fs.WriteString("Answer:" + "\n" + RESP + "\n")
			if err != nil {
				panic(err)
			}
			fs.Close()
		}

		// Speak all the response RESP using the "say" command
		if speak == 1 {

			fmt.Println("speaking")
			go func() {
				switch runtime.GOOS {
				case "linux", "darwin":
					cmd := exec.Command("say", RESP)
					err = cmd.Run()
					if err != nil {
						fmt.Println(err)
					}
				case "windows":
					// to do
					_ = 1 + 1

				}

			}()
		}

	}
}

func scrollUp(textView *tview.TextView) {
	row, _ := textView.GetScrollOffset()
	if row > 0 {
		textView.ScrollTo(row-1, 0)
	}
}

func scrollDown(textView *tview.TextView) {
	row, _ := textView.GetScrollOffset()
	textView.ScrollTo(row+1, 0)
}

func printer(colour tcell.Color, context string, history bool) {
	app := tview.NewApplication()
	flex := tview.NewFlex()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true).
		SetTextColor(colour)

	flex.AddItem(tview.NewTextView(), 0, 1, false).AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView(), 0, 1, false).
		AddItem(textView, 0, 6, true).
		AddItem(tview.NewTextView(), 0, 1, false), 0, 8, false).
		AddItem(tview.NewTextView(), 0, 1, false)

	fmt.Fprintf(textView, context)
	if history {
		textView.ScrollToEnd()
	}

	// Handle 'jkgGq'
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			app.Stop()
			//	case tcell.KeyUp: // maybe use for last response
			//		scrollUp(textView)
			//	case tcell.KeyDown:
		case tcell.KeyRune:
			switch event.Rune() {
			case 'k':
				scrollUp(textView)
			case 'j':
				scrollDown(textView)
			case 'g':
				textView.ScrollToBeginning()
			case 'G':
				textView.ScrollToEnd()
			case 'q':
				app.Stop()
			}
		}
		return event
	})

	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}

}
package main

import (
	//	"context"
	_ "embed"
	"fmt"
	//	"github.com/atotto/clipboard"
	//	"github.com/creack/pty"
	//	"github.com/gdamore/tcell/v2"
	"github.com/go-rod/rod"
	//	"github.com/go-rod/rod/lib/launcher"
	//	"github.com/go-rod/rod/lib/utils"
	"github.com/go-rod/stealth"
	//"github.com/google/uuid"
	//	"github.com/manifoldco/promptui"
	//	"github.com/peterh/liner"
	//	"github.com/rivo/tview"
	//	openai "github.com/sashabaranov/go-openai"
	//"github.com/tidwall/gjson"
	//	"github.com/tidwall/sjson"
	//	"golang.org/x/crypto/ssh/terminal"
	//	"io"
	//	"io/ioutil"
	//	"log"
	//	"os"
	//	"os/exec"
	//	"os/signal"
	//	"runtime"
	//"strconv"
	"strings"
	//	"syscall"
	"time"
)

var browser *rod.Browser

// Set up client of Bard (Rod version)
var page_bard *rod.Page
var relogin_bard = true
var channel_bard chan string

func Bard() {
	channel_bard = make(chan string)
	defer func() {
		if err := recover(); err != nil {
			relogin_bard = true
		}
	}()
	page_bard = stealth.MustPage(browser)
	page_bard.MustNavigate("https://bard.google.com")

	for i := 1; i <= 30; i++ {
		if page_bard.MustHasX("//textarea[@id='mat-input-0']") {
			relogin_bard = false
			break
		}
		// Check "I'm not a robot"
		info := page_bard.MustInfo()
		if strings.HasPrefix(info.URL, "https://google.com/sorry") {
			relogin_bard = true
			break
		}
		// Check "Sign in"
		if page_bard.MustHasX("//a[contains(text(), 'Sign in')]") {
			relogin_bard = true
			break
		}
		// Check "You've been signed out"
		if page_bard.MustHasX("//h1[contains(text(), 've been signed out')]") {
			relogin_bard = true
			break
		}

		time.Sleep(time.Second)
	}
	if relogin_bard == true {
		sprint("✘ Bard")
	}
	if relogin_bard == false {
		sprint("✔ Bard")
		for {
			select {
			case question := <-channel_bard:
				//page_bard.MustActivate()
				page_bard.MustElementX("//textarea[@id='mat-input-0']").MustWaitVisible().MustInput(question)
				page_bard.MustElementX("//button[@mattooltip='Submit']").MustClick()
				fmt.Println("Bard generating...")
				//page_bard.MustActivate()
				//if role == ".all" {
				//	channel_bard <- "click_bard"
				//}
				// wait generated icon
				var generated_icon_appear = false
				for i := 1; i <= 60; i++ {
					if page_bard.MustHasX("//img[contains(@src, 'https://www.gstatic.com/lamda/images/sparkle_resting_v2_1ff6f6a71f2d298b1a31.gif')]") {
						generated_icon_appear = true
						break
					}
					time.Sleep(1 * time.Second)
				}
				if generated_icon_appear == true {
					img := page_bard.MustElementX("//img[contains(@src, 'https://www.gstatic.com/lamda/images/sparkle_resting_v2_1ff6f6a71f2d298b1a31.gif')][last()]").MustWaitVisible()
					response := img.MustElementX("parent::div/parent::div").MustWaitVisible()
					answer := response.MustText()
					channel_bard <- answer
				} else {
					channel_bard <- "✘✘  Bard, Please check the internet connection and verify login status."
					relogin_bard = true

				}
			}
		}
	}

}

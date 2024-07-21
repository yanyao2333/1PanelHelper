package main

import (
	"1PanelHelper/notify"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"os"
	"strconv"
)

func main() {
	p := PanelApiStruct{
		BaseUrl: os.Getenv("PANEL_BASE_URL"),
	}
	err := p.LoginWithout2FA(os.Getenv("PANEL_USERNAME"), os.Getenv("PANEL_PASSWORD"), os.Getenv("PANEL_ENTRANCE_CODE"))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = p.GetCronjobList(SearchRequest{
		Page:     1,
		PageSize: 10,
		Order:    "null",
		OrderBy:  "created_at",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Printf("%v", list)
	records, err := p.GetCronjobRecords(CronjobRecordRequest{
		CronjobID: 1,
		Page:      1,
		PageSize:  8,
		Days:      3,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v", records)

	gotify := notify.GotifyStruct{
		ServerRootUrl: os.Getenv("GOTIFY_BASE_URL"),
		Token:         os.Getenv("GOTIFY_APP_APIKEY"),
	}
	gotify.Priority, _ = strconv.Atoi(os.Getenv("GOTIFY_PRIORITY"))
	_ = gotify.PushNotify("test", "test")
}

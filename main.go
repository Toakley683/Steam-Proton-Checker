package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	SteamAPI "github.com/Toakley683/GoLang-SteamAPI"
)

func main() {

	APIKey, err := GetAPIKey()
	if err != nil {
		log.Panicln("Could not get API Key")
	}

	SteamAPI.SetSteamAPIContext(APIKey)

	Toakley682 := SteamAPI.ClientInformation{
		SteamID64: "76561198170087194",
	}

	var MaxAttempts int = 25
	var Attempts int = 0
	var Games *SteamAPI.PlayerGamesList

	for Games == nil {
		Attempts++
		G, err := Toakley682.GetOwnedGames()

		if Attempts > MaxAttempts {
			log.Panicln("Could not get Game List after", MaxAttempts, "attempts")
		}

		time.Sleep(time.Millisecond * 500)

		if err != nil {
			log.Println("Could not get Game List, Retrying..")
			continue
		}

		Games = G

	}

	type GameReport []struct {
		ID            int         `json:"id"`
		AppID         int         `json:"appId"`
		Timestamp     int         `json:"timestamp"`
		Rating        string      `json:"rating"`
		Notes         string      `json:"notes"`
		Os            string      `json:"os"`
		GpuDriver     string      `json:"gpuDriver"`
		Specs         interface{} `json:"specs"`
		ProtonVersion string      `json:"protonVersion"`
	}

	ReportInfo := struct {
		PlatinumMedals    map[int]SteamAPI.PlayerGame
		GoldMedals        map[int]SteamAPI.PlayerGame
		SilverMedals      map[int]SteamAPI.PlayerGame
		BronzeMedals      map[int]SteamAPI.PlayerGame
		BorkedMedals      map[int]SteamAPI.PlayerGame
		UnavailableMedals map[int]SteamAPI.PlayerGame
	}{
		PlatinumMedals:    map[int]SteamAPI.PlayerGame{},
		GoldMedals:        map[int]SteamAPI.PlayerGame{},
		SilverMedals:      map[int]SteamAPI.PlayerGame{},
		BronzeMedals:      map[int]SteamAPI.PlayerGame{},
		BorkedMedals:      map[int]SteamAPI.PlayerGame{},
		UnavailableMedals: map[int]SteamAPI.PlayerGame{},
	}

	var MedalLookup = map[string]int{
		"Borked":   0,
		"Bronze":   1,
		"Silver":   2,
		"Gold":     3,
		"Platinum": 4,
	}

	for _, game := range Games.Data.Games {
		appID := strconv.Itoa(game.AppID)

		Response, err := http.Get("https://protondb.max-p.me/games/" + appID + "/reports/")
		if err != nil {
			fmt.Println(err)
			continue
		}

		if Response.StatusCode != 200 {
			fmt.Println("Status was not 200, Status was:", Response.Status)
			continue
		}

		Data, err := io.ReadAll(Response.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}

		gameReport := GameReport{}

		err = json.Unmarshal(Data, &gameReport)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if len(gameReport) <= 0 {
			fmt.Println("AppID:", game.AppID, "Rating:", "N/A")
			ReportInfo.UnavailableMedals[len(ReportInfo.UnavailableMedals)] = game
			continue
		}

		var FinalReport string

		for _, v := range gameReport {

			if FinalReport == "" {
				FinalReport = v.Rating
			}

			if MedalLookup[v.Rating] > MedalLookup[FinalReport] {
				FinalReport = v.Rating
			}
		}

		Report := gameReport[0].Rating

		switch Report {
		case "Platinum":
			ReportInfo.PlatinumMedals[len(ReportInfo.PlatinumMedals)] = game
		case "Gold":
			ReportInfo.GoldMedals[len(ReportInfo.GoldMedals)] = game
		case "Silver":
			ReportInfo.SilverMedals[len(ReportInfo.SilverMedals)] = game
		case "Bronze":
			ReportInfo.BronzeMedals[len(ReportInfo.BronzeMedals)] = game
		case "Borked":
			ReportInfo.BorkedMedals[len(ReportInfo.BorkedMedals)] = game
		default:
			ReportInfo.UnavailableMedals[len(ReportInfo.UnavailableMedals)] = game
		}

		fmt.Println("AppID:", strconv.Itoa(game.AppID), "Rating:", Report)

	}

	GameTotal := len(Games.Data.Games)
	GameTotalString := strconv.Itoa(GameTotal)

	fmt.Println("Game Report: [Total " + GameTotalString + "]")
	fmt.Println("Platinum:", strconv.Itoa(len(ReportInfo.PlatinumMedals))+"/"+GameTotalString)
	fmt.Println("Gold:", strconv.Itoa(len(ReportInfo.GoldMedals))+"/"+GameTotalString)
	fmt.Println("Silver:", strconv.Itoa(len(ReportInfo.SilverMedals))+"/"+GameTotalString)
	fmt.Println("Bronze:", strconv.Itoa(len(ReportInfo.BronzeMedals))+"/"+GameTotalString)
	fmt.Println("Borked:", strconv.Itoa(len(ReportInfo.BorkedMedals))+"/"+GameTotalString)
	fmt.Println("No Reports:", strconv.Itoa(len(ReportInfo.UnavailableMedals))+"/"+GameTotalString)

}

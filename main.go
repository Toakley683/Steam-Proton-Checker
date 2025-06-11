package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/roblillack/spot"
	"github.com/roblillack/spot/ui"

	SteamAPI "github.com/Toakley683/GoLang-SteamAPI"
)

func main() {

	ui.Init()

	APIKey, err := GetAPIKey()
	if err != nil {
		log.Panicln("Could not get API Key")
	}

	Context := SteamAPI.SetSteamAPIContext(APIKey)

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

	Apps, err := Context.GetAppList()
	if err != nil {
		log.Panicln("Could not get App List")
	}

	AppNameLookup := map[int]string{}

	for i := 0; i < len(Apps.Data.Apps); i++ {
		Data := Apps.Data.Apps[i]
		AppNameLookup[Data.AppID] = Data.Name
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
		"No Record": 0,
		"Borked":    1,
		"Bronze":    2,
		"Silver":    3,
		"Gold":      4,
		"Platinum":  5,
	}

	updateFuncRunning := false

	spot.MountFn(func(ctx *spot.RenderContext) spot.Component {

		progress, setProgress := spot.UseState[float64](ctx, 0)

		platinumMedals, setPlatinumMedals := spot.UseState(ctx, []string{})
		goldMedals, setGoldMedals := spot.UseState(ctx, []string{})
		silverMedals, setSilverMedals := spot.UseState(ctx, []string{})
		bronzeMedals, setBronzeMedals := spot.UseState(ctx, []string{})
		borkedMedals, setBorkedMedals := spot.UseState(ctx, []string{})
		unavailableMedals, setUnavailableMedals := spot.UseState(ctx, []string{})

		onUpdate := func() {

			if updateFuncRunning {
				return
			}
			updateFuncRunning = true

			setPlatinumMedals([]string{})
			setGoldMedals([]string{})
			setSilverMedals([]string{})
			setBronzeMedals([]string{})
			setBorkedMedals([]string{})
			setUnavailableMedals([]string{})

			ReportInfo = struct {
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

			for Index, game := range Games.Data.Games {
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

				if AppNameLookup[game.AppID] == "" {
					AppNameLookup[game.AppID] = "[AppID: " + strconv.Itoa(game.AppID) + "]"
				}

				if len(gameReport) <= 0 {
					fmt.Println("App:", AppNameLookup[game.AppID], "Rating:", "N/A")
					setProgress(float64(Index))

					ReportInfo.UnavailableMedals[len(ReportInfo.UnavailableMedals)] = game

					Medals := make([]string, len(ReportInfo.UnavailableMedals))
					for I, v := range ReportInfo.UnavailableMedals {
						Medals[I] = AppNameLookup[v.AppID]
					}

					setUnavailableMedals(Medals)
					continue
				}

				var FinalReport string

				for _, v := range gameReport {

					if FinalReport == "" {
						FinalReport = v.Rating
					}

					if MedalLookup[v.Rating] > MedalLookup[FinalReport] {
						fmt.Println("Exchanged", v.Rating, "for", FinalReport)
						FinalReport = v.Rating
					}
				}

				switch FinalReport {
				case "Platinum":
					ReportInfo.PlatinumMedals[len(ReportInfo.PlatinumMedals)] = game

					Medals := make([]string, len(ReportInfo.PlatinumMedals))
					for I, v := range ReportInfo.PlatinumMedals {
						Medals[I] = AppNameLookup[v.AppID]
					}

					setPlatinumMedals(Medals)
				case "Gold":
					ReportInfo.GoldMedals[len(ReportInfo.GoldMedals)] = game

					Medals := make([]string, len(ReportInfo.GoldMedals))
					for I, v := range ReportInfo.GoldMedals {
						Medals[I] = AppNameLookup[v.AppID]
					}

					setGoldMedals(Medals)
				case "Silver":
					ReportInfo.SilverMedals[len(ReportInfo.SilverMedals)] = game

					Medals := make([]string, len(ReportInfo.SilverMedals))
					for I, v := range ReportInfo.SilverMedals {
						Medals[I] = AppNameLookup[v.AppID]
					}

					setSilverMedals(Medals)
				case "Bronze":
					ReportInfo.BronzeMedals[len(ReportInfo.BronzeMedals)] = game

					Medals := make([]string, len(ReportInfo.BronzeMedals))
					for I, v := range ReportInfo.BronzeMedals {
						Medals[I] = AppNameLookup[v.AppID]
					}

					setBronzeMedals(Medals)
				case "Borked":
					ReportInfo.BorkedMedals[len(ReportInfo.BorkedMedals)] = game

					Medals := make([]string, len(ReportInfo.BorkedMedals))
					for I, v := range ReportInfo.BorkedMedals {
						Medals[I] = AppNameLookup[v.AppID]
					}

					setBorkedMedals(Medals)
				}

				setProgress(float64(Index))

				fmt.Println("App:", AppNameLookup[game.AppID], "Rating:", FinalReport)

			}

			updateFuncRunning = false

		}

		WindowWidth := 1280
		WindowHeight := 900

		windowChildren := make([]spot.Component, 2+(len(MedalLookup)*3))

		for medal, v := range MedalLookup {

			var values []string

			switch v {
			case 0:
				values = unavailableMedals
			case 1:
				values = borkedMedals
			case 2:
				values = bronzeMedals
			case 3:
				values = silverMedals
			case 4:
				values = goldMedals
			case 5:
				values = platinumMedals
			default:
				values = []string{}
			}

			W := (WindowWidth - 50) / len(MedalLookup)

			Index := v * 3

			windowChildren[3+(Index-1)] = &ui.ListBox{
				X: 25 + (W * v), Y: 100, Width: W, Height: WindowHeight - (100 + 100),
				Values: values,
			}

			windowChildren[3+(Index-2)] = &ui.TextField{
				X: 25 + (W * v), Y: 75, Width: W, Height: 25,
				Value: medal,
			}

			windowChildren[3+(Index-3)] = &ui.TextView{
				X: 25 + (W * v), Y: WindowHeight - (100), Width: W, Height: 25,
				Text: "Total: " + strconv.Itoa(len(values)),
			}
		}

		windowChildren[(len(MedalLookup) * 3)] = &ui.Button{
			X: 25, Y: WindowHeight - 50, Width: WindowWidth - 50, Height: 25,
			Title: "Refresh Games",
			OnClick: func() {
				setProgress(0)
				go onUpdate()
			},
		}

		windowChildren[(len(MedalLookup)*3)+1] = &ui.ProgressBar{
			X: 25, Y: 25, Width: WindowWidth - 50, Height: 25,
			Min:   0,
			Max:   float64(len(Games.Data.Games)),
			Value: progress,
		}

		Wind := &ui.Window{
			Title:    "Proton Graph",
			Width:    WindowWidth,
			Height:   WindowHeight,
			Children: windowChildren,
		}

		return Wind
	})

	ui.Run()
}

/*

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
*/

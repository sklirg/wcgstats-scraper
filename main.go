package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/redis.v5"
)

// Challenge API - Challenge ID 8183
const challenge_api = "https://www.worldcommunitygrid.org/team/challenge/viewTeamChallenge.do"

// + ?challengeId=8183&xml=true

// https://www.worldcommunitygrid.org/team/viewTeamInfo.do?teamId=3KWRNGFL72&xml=true
const teaminfo_api = "https://www.worldcommunitygrid.org/team/viewTeamInfo.do"

// + ?teamId=3KWRNGFL72&xml=true

// Stats
// https://www.worldcommunitygrid.org/team/viewTeamStatHistory.do?teamId=3KWRNGFL72&xml=true
const teamstats_api = "https://www.worldcommunitygrid.org/team/viewTeamStatHistory.do"

// + ?teamId=3KWRNGFL72&xml=true

func main() {
	team_id := os.Getenv("WCGSTATS_SCRAPER_TEAM")
	if team_id == "" {
		log.Error("No team specified. Please set the env var WCGSTATS_SCRAPER_TEAM to the team ID you want to scrape.")
		return
	}
	log.WithField("team", team_id).Info("Starting scraper.")

	GetTeamStatsHistory(team_id)
}

// @ToDo: implement
func GetTeamInfo(team_id string) {}

func GetTeamStatsHistory(team_id string) {
	log.WithFields(log.Fields{
		"team_id": team_id,
	}).Infof("Gathering team stats for %v", team_id)

	var Url *url.URL
	Url, _ = url.Parse(teamstats_api)

	query_params := url.Values{}
	query_params.Add("teamId", team_id)
	query_params.Add("xml", "true")
	Url.RawQuery = query_params.Encode()

	resp, err := http.Get(Url.String())

	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"url":     Url,
			"team_id": team_id,
		}).Error("An error occured while gathering team stats.")
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"body":    body,
			"team_id": team_id,
		}).Error("An error occured while reading response body.")
		return
	}

	v := StatisticsHistory{}

	log.WithFields(log.Fields{
		"status": resp.StatusCode,
	}).Info("Response status")

	err = xml.Unmarshal(body, &v)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"team_id": team_id,
			"data":    v,
		}).Error("An error occured while unmarshaling response body.")
		return
	}

	log.WithFields(log.Fields{
		"data": v,
	}).Debug("Response struct")

	j, err := json.Marshal(v)

	log.WithFields(log.Fields{
		"json": string(j),
	}).Debug("Marshaled JSON")

	PostRedisData(team_id, v)
}

func PostRedisData(team_id string, data StatisticsHistory) {
	redis_host := os.Getenv("WCGSTATS_SCRAPER_REDIS_HOST")
	if redis_host == "" {
		redis_host = "localhost"
	}
	log.WithFields(log.Fields{
		"team_id":    team_id,
		"redis_host": redis_host,
	}).Info("Posting stats to redis")
	client := redis.NewClient(&redis.Options{
		Addr:     redis_host + ":6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	client.HSet("teams", team_id, "")

	log.WithFields(log.Fields{
		"data": data.DailyStatisticsTotals,
	}).Debugf("Values to insert.")

	for i := range data.DailyStatisticsTotals {
		j, err := json.Marshal(data.DailyStatisticsTotals[i])
		log.WithFields(log.Fields{
			"err":  err,
			"data": data.DailyStatisticsTotals[i],
			"json": string(j),
		}).Debug("Value to insert")

		res := client.HSet(team_id, string(data.DailyStatisticsTotals[i].Date), string(j))
		if res.Err() != nil {
			log.WithFields(log.Fields{
				"value": res.Val(),
				"error": res.Err(),
				"i":     i,
			}).Error("Response from request.")
		} else {
			log.WithFields(log.Fields{
				"value": res.Val(),
				"error": res.Err(),
				"i":     i,
			}).Info("Response from request.")
		}
	}
}

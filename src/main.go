package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

type Channel struct {
	Channel string `json:"channel"`
}

type Message struct {
	Id      int    `json:"id"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type ActiveStruct struct {
	MLU ActiveMluStruct `json:"MLU"`
}

type ActiveMluStruct struct {
	SCH string `json:"SCH"`
	SCA string `json:"SCA"`
	CR  string `json:"CR"`
}

type ServerStruct struct {
	Host      string `json:"host"`
	Token     string `json:"token"`
	Home      string `json:"home"`
	Away      string `json:"away"`
	StartTime string `json:"startTime"`
	GsmId     string `json:"gsmId"`
}

func main() {

	// 读取配置
	log.Println("running")
	serverBuf, err := ioutil.ReadFile("server.json")
	if err != nil {
		log.Println("config error", err)
	}

	servers := make([]ServerStruct, 0, 1024)
	err = json.Unmarshal(serverBuf, &servers)
	if err != nil {
		log.Println("config error")
	}

	var wg sync.WaitGroup

	// 开启SERVER
	for _, v := range servers {

		// 校验时间
		loc, _ := time.LoadLocation("Local")
		tm, _ := time.ParseInLocation("2006-01-02 15:04:05", v.StartTime, loc)
		startTime := tm.Unix()
		nowTime := time.Now().Unix()
		endTime := startTime + 7200
		log.Println(startTime, nowTime, endTime)
		if startTime <= nowTime && nowTime < endTime {

			log.Println("match:", v.Home, v.Away)
			wg.Add(1)

			go func(host string, token string, home string, away string, wg *sync.WaitGroup, gsmId string) {
				getMatch(host, token, home, away, endTime, wg, gsmId)
			}(v.Host, v.Token, v.Home, v.Away, &wg, v.GsmId)
		} else if startTime > nowTime {
			log.Println("match:", v.Home, v.Away)

			wg.Add(1)
			go func(host string, token string, home string, away string, wg *sync.WaitGroup, gsmId string) {
				timeInterval := startTime - nowTime
				timer := time.NewTicker(time.Duration(timeInterval) * time.Second)
				select {
				case <-timer.C:
					getMatch(host, token, home, away, endTime, wg, gsmId)
				}
			}(v.Host, v.Token, v.Home, v.Away, &wg, v.GsmId)
		} else {
			log.Println("match over:", v.Home, v.Away)
		}
	}

	wg.Wait()
	log.Println("done")
}

func getMatch(host string, token string, Home string, Away string, EndTime int64, wg *sync.WaitGroup, gsmId string) {

	protocol := "ws://"

	url := protocol + host + "/socket.io/?token=" + token + "&EIO=3&transport=websocket"

	t := transport.GetDefaultWebsocketTransport()
	count := 0

	c, err := gosocketio.Dial(url, t)
	if err != nil {
		log.Println("connect error:", err)
	}
	activeMP := &ActiveStruct{}
	nowScore := ""
	newScore := ""
	title := Home + " VS " + Away

	err = c.On("message", func(h *gosocketio.Channel, args interface{}) {

		if mapData, ok := args.(map[string]interface{}); ok {
			if activeMsg, ok := mapData["ActiveMQMessage"]; ok {
				err := json.Unmarshal([]byte(activeMsg.(string)), activeMP)
				if activeMP.MLU.CR == "false" {
					// 结束协程
					nowTime := time.Now().Unix()
					if nowTime >= EndTime {
						log.Println("over")
						wg.Done()
					}
				}

				if err == nil {
					newScore = activeMP.MLU.SCH + ":" + activeMP.MLU.SCA
					if newScore != nowScore {
						nowScore = newScore
						send(title + newScore)
						log.Println(title, activeMP.MLU.SCH+":"+activeMP.MLU.SCA)
					}
					count = 0
					file, err := os.OpenFile(gsmId+".log", os.O_CREATE|os.O_APPEND, 0666)
					if err != nil {
						log.Fatal(err)
					}

					buf := bytes.NewBuffer([]byte{})
					fmt.Fprintln(buf, args)
					file.Write(buf.Bytes())
					file.Close()
				} else {
					log.Println(err)
				}
			}
		}
	})

	if err != nil {
		log.Fatal(err)
	}

	err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Println("Connected")
	})

	if err != nil {
		log.Fatal(err)
	}

	err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		log.Println("Disconnection")
		count = 5
	})

	if err != nil {
		log.Fatal(err)
	}

	c.On(gosocketio.OnError, func(h *gosocketio.Channel) {
		log.Println("err")
	})

	// 设置定时器 定时查看是否有消息接收 如果没有 关闭连接 根据时间确定是否需要重启
	timer := time.NewTicker(30 * time.Second)
	reloadFlag := 0
	stopTimer := 0
	for {
		if stopTimer == 1 {
			break
		}
		select {
		case <-timer.C:
			log.Println("count:", count)
			count++
			if count >= 5 {
				nowTime := time.Now().Unix()
				c.Close()
				stopTimer = 1
				if nowTime >= EndTime {
					log.Println("end time done")
					wg.Done()
				} else {
					reloadFlag = 1
					log.Println("reload")
				}
			}
		}
	}

	if reloadFlag == 1 {
		go getMatch(host, token, Home, Away, EndTime, wg, gsmId)
	}
}

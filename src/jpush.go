package main

import (
	"fmt"

	"github.com/ylywyn/jpush-api-go-client"
)

const (
	appKey = "" // jpush appkey
	secret = "" // jpush secret
)

func send(content string) {

	//Platform
	var pf jpushclient.Platform
	//pf.Add(jpushclient.ANDROID)
	pf.Add(jpushclient.IOS)
	//pf.Add(jpushclient.WINPHONE)
	//pf.All()

	//Audience
	var ad jpushclient.Audience
	// 推送的组
	s := []string{"groupName"}
	ad.SetTag(s)
	ad.SetAlias(s)
	//ad.SetID(s)
	//ad.All()

	//Notice
	var notice jpushclient.Notice
	notice.SetAlert(content)
	//notice.SetAndroidNotice(&jpushclient.AndroidNotice{Alert: "AndroidNotice"})
	notice.SetIOSNotice(&jpushclient.IOSNotice{Alert: content})
	//notice.SetWinPhoneNotice(&jpushclient.WinPhoneNotice{Alert: "WinPhoneNotice"})

	//var msg jpushclient.Message
	//msg.Title = "比分:"
	//msg.Content = "1:0"

	options := &jpushclient.Option{}
	options.ApnsProduction = true
	payload := jpushclient.NewPushPayLoad()
	payload.SetOptions(options)
	payload.SetPlatform(&pf)
	payload.SetAudience(&ad)
	//payload.SetMessage(&msg)
	payload.SetNotice(&notice)

	bytes, _ := payload.ToBytes()

	//push
	c := jpushclient.NewPushClient(secret, appKey)
	str, err := c.Send(bytes)
	if err != nil {
		fmt.Printf("err:%s", err.Error())
	} else {
		fmt.Printf("ok:%s", str)
	}
}

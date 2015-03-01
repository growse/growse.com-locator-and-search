package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("auth")
		ok := false
		log.Printf("Cookie: %v", cookie)
		log.Println(err)
		if err == nil && cookie != nil {
			_, ok = validateCookie(cookie, configuration.CookieSeed)
		}
		log.Printf("OK: %v", ok)
		if err != nil || !ok {
			url := oAuthConf.AuthCodeURL(c.Request.URL.String())
			c.Set("redirecturl", url)
		}
	}
}

func OauthCallback(c *gin.Context) {
	c.Request.ParseForm()
	authCode := c.Request.Form.Get("code")
	state := c.Request.Form.Get("state")

	tok, err := oAuthConf.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatal(err)
	}
	client := oAuthConf.Client(oauth2.NoContext, tok)

	profileUrl := "https://www.googleapis.com/userinfo/v2/me"
	resp, err := client.Get(profileUrl)

	if err != nil {
		log.Fatal(err)
	}
	responsebytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	type responsestruct struct {
		Id            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		Hd            string `json:"hd"`
	}

	var responseObject responsestruct
	json.Unmarshal(responsebytes, &responseObject)
	log.Println(responseObject)
	if responseObject.Hd == "growse.com" {
		cookie := &http.Cookie{
			Name:     "auth",
			Value:    signedCookieValue(configuration.CookieSeed, "auth", responseObject.Email),
			Path:     "/",
			Domain:   configuration.Domain,
			HttpOnly: !configuration.Production,
			Secure:   configuration.Production,
			Expires:  time.Now().Add(time.Minute * 15),
		}
		log.Printf("Setting cookie: %v\n", cookie)
		http.SetCookie(c.Writer, cookie)
		c.JSON(200, "Success "+state)
	} else {
		c.JSON(500, string(responsebytes))
	}
}

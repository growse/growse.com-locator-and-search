package main

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("auth")
		ok := false
		cookieContent := ""
		log.Printf("Supplied auth cookie: %v", cookie)
		if err == nil && cookie != nil {
			cookieContent, ok = validateCookie(cookie, configuration.CookieSeed)
		}
		log.Printf("Cookie valid?: %v", ok)
		if err != nil || !ok {
			url := oAuthConf.AuthCodeURL(c.Request.URL.String())
			c.Redirect(302, url)
			c.Abort()
			return
		}
		log.Printf("Cookie contents: %v", cookieContent)
		c.Next()

	}
}

func OauthCallback(c *gin.Context) {
	c.Request.ParseForm()
	authCode := c.Request.Form.Get("code")
	state := c.Request.Form.Get("state")
	oauthContext := context.Background()
	tok, err := oAuthConf.Exchange(oauthContext, authCode)
	if err != nil {
		c.AbortWithError(500, err)
	}
	client := oAuthConf.Client(oauthContext, tok)

	profileUrl := "https://www.googleapis.com/userinfo/v2/me"
	resp, err := client.Get(profileUrl)

	if err != nil {
		c.AbortWithError(500, err)
	}
	responsebytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithError(500, err)
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
	if responseObject.Hd == "growse.com" || responseObject.Email == "growse@gmail.com" {
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

		if state != "" {
			c.Redirect(302, state)
		} else {
			c.String(200, "Success")
		}
	} else {
		c.JSON(401, string(responsebytes))
	}
}

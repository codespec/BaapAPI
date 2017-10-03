package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/BaapAPI/structs"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

var (
	gooleconf    *oauth2.Config
	facebookconf *oauth2.Config
	procred      ProvideCre
)

// Credentials which stores id and key.
type Credentials struct {
	Cid     string `json:"client_id"`
	Csecret string `json:"client_secret"`
}

//ProvideCre which store different id and key for provider
type ProvideCre map[string]Credentials

// RandToken generates a random @l length token.
func RandToken(l int) string {
	b := make([]byte, l)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func getLoginURL(state, provider string) string {
	if provider == "facebook" {
		return facebookconf.AuthCodeURL(state)
	} else if provider == "google" {
		return gooleconf.AuthCodeURL(state)
	} else {
		return ""
	}

}

func init() {
	file, err := ioutil.ReadFile("./client_secret.json")
	if err != nil {
		log.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &procred)

	gooleconf = &oauth2.Config{
		ClientID:     procred["google"].Cid,
		ClientSecret: procred["google"].Csecret,
		RedirectURL:  "http://127.0.0.1:9090/google/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}

	facebookconf = &oauth2.Config{
		ClientID:     procred["facebook"].Cid,
		ClientSecret: procred["facebook"].Csecret,
		RedirectURL:  "http://localhost:9090/facebook/auth",
		Scopes:       []string{"public_profile"},
		Endpoint:     facebook.Endpoint,
	}

}

// IndexHandler handels /.
func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{})
}

// GoogleAuthHandler handles authentication of a user and initiates a session.
func GoogleAuthHandler(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)
	retrievedState := session.Get("state")
	queryState := c.Request.URL.Query().Get("state")
	if retrievedState != queryState {
		log.Printf("Invalid session state: retrieved: %s; Param: %s", retrievedState, queryState)
		c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "Invalid session state."})
		return
	}
	code := c.Request.URL.Query().Get("code")
	tok, err := gooleconf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Login failed. Please try again."})
		return
	}

	client := gooleconf.Client(oauth2.NoContext, tok)
	userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer userinfo.Body.Close()
	data, _ := ioutil.ReadAll(userinfo.Body)
	u := structs.GoogleUser{}
	if err = json.Unmarshal(data, &u); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error marshalling response. Please try agian."})
		return
	}
	session.Set("user-id", u.Email)
	err = session.Save()
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error while saving session. Please try again."})
		return
	}
	seen := false
	/* 	db := database.MongoDBConnection{}
	   	if _, mongoErr := db.LoadUser(u.Email); mongoErr == nil {
	   		seen = true
	   	} else {
	   		err = db.SaveUser(&u)
	   		if err != nil {
	   			log.Println(err)
	   			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error while saving user. Please try again."})
	   			return
	   		}
	   	} */
	c.HTML(http.StatusOK, "hello.tmpl", gin.H{"info": "google user:" + u.Email, "seen": seen})
}

// FaceBookAuthHandler handles authentication of a user and initiates a session.
func FaceBookAuthHandler(c *gin.Context) {
	retrievedState := c.Request.FormValue("state")
	queryState := c.Request.URL.Query().Get("state")
	if retrievedState != queryState {
		log.Printf("Invalid session state: retrieved: %s; Param: %s", retrievedState, queryState)
		c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "Invalid session state."})
		return
	}
	code := c.Request.URL.Query().Get("code")
	tok, err := facebookconf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Login failed. Please try again."})
		return
	}

	client := facebookconf.Client(oauth2.NoContext, tok)
	userinfo, err := client.Get("https://graph.facebook.com/me?")
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer userinfo.Body.Close()
	data, _ := ioutil.ReadAll(userinfo.Body)
	u := structs.FaceBookUser{}
	if err = json.Unmarshal(data, &u); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error marshalling response. Please try agian."})
		return
	}
	//session.Set("user-id", u.Email)
	/* err = session.Save()
	if err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error while saving session. Please try again."})
		return
	} */
	seen := false
	/* 	db := database.MongoDBConnection{}
	   	if _, mongoErr := db.LoadUser(u.Email); mongoErr == nil {
	   		seen = true
	   	} else {
	   		err = db.SaveUser(&u)
	   		if err != nil {
	   			log.Println(err)
	   			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error while saving user. Please try again."})
	   			return
	   		}
	   	} */
	c.HTML(http.StatusOK, "hello.tmpl", gin.H{"info": "facebook user:" + u.Name, "seen": seen})
}

// LoginHandler handles the login procedure.
func LoginHandler(c *gin.Context) {
	state := RandToken(32)
	session := sessions.Default(c)
	session.Set("state", state)
	log.Printf("Stored session: %v\n", state)
	session.Save()
	glink := getLoginURL(state, "google")
	flink := getLoginURL(state, "facebook")
	c.HTML(http.StatusOK, "auth.tmpl", gin.H{"glink": glink, "flink": flink})
}

// FieldHandler is a rudementary handler for logged in users.
func FieldHandler(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user-id")
	c.HTML(http.StatusOK, "field.tmpl", gin.H{"user": userID})
}

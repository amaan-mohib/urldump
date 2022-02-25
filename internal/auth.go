package internal

import (
	"cli/models"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/viper"
)

func ValidateToken(accessToken string) (bool, error) {
	url := "http://localhost:3000/validate-token"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("HTTP error: %s", err)
		return false, err
	}

	status := res.StatusCode

	if status == 200 {
		return true, nil
	} else {
		return false, nil
	}
}

func AuthorizeUser(clientID string, audience string, clientSecret string, authDomain string, redirectURL string) {
	CodeVerifier, _ := cv.CreateCodeVerifier()

	codeChallenge := CodeVerifier.CodeChallengeS256()

	authorizationURL := fmt.Sprintf("https://%s/authorize?"+
		"&response_type=code&client_id=%s"+"&audience=%s"+
		"&scope=openid profile email offline_access"+
		"&code_challenge=%s"+
		"&code_challenge_method=S256&redirect_uri=%s",
		authDomain, clientID, audience, codeChallenge, redirectURL)

	server := &http.Server{Addr: redirectURL}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			fmt.Println("Url Param 'code' is missing")
			io.WriteString(w, "Error: could not find 'code' URL parameter\n")

			// close the HTTP server and return
			cleanup(server)
			return
		}

		// trade the authorization code and the code verifier for an access token
		codeVerifier := CodeVerifier.String()
		token, err := getAccessToken(clientID, clientSecret, codeVerifier, code, redirectURL)
		if err != nil {
			fmt.Println("could not get access token")
			io.WriteString(w, "Error: could not retrieve access token\n")

			// close the HTTP server and return
			cleanup(server)
			return
		}
		user, err := getUser(token.AccessToken)
		if err != nil {
			fmt.Println("could not get user info")
			io.WriteString(w, "Error: could not retrieve user info\n")

			// close the HTTP server and return
			cleanup(server)
			return
		}

		viper.Set("Tokens", token)
		viper.Set("user", user)
		err = viper.WriteConfig()

		if err != nil {
			fmt.Println(err)
			fmt.Println("could not write config file")
			io.WriteString(w, "Error: could not store access token\n")

			// close the HTTP server and return
			cleanup(server)
			return
		}

		// return an indication of success to the caller
		io.WriteString(w, `
		<html>
			<body>
				<h1>Login successful!</h1>
				<h2>You can close this window and return to the urldump CLI.</h2>
			</body>
		</html>`)

		fmt.Println("Successfully logged in.")

		// close the HTTP server
		cleanup(server)

	})
	// parse the redirect URL for the port number
	u, err := url.Parse(redirectURL)
	if err != nil {
		fmt.Printf("bad redirect URL: %s\n", err)
		os.Exit(1)
	}

	// set up a listener on the redirect port
	port := fmt.Sprintf(":%s", u.Port())
	l, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("can't listen to port %s: %s\n", port, err)
		os.Exit(1)
	}

	// open a browser window to the authorizationURL
	err = open.Start(authorizationURL)
	if err != nil {
		fmt.Printf("can't open browser to URL %s: %s\n", authorizationURL, err)
		os.Exit(1)
	}

	// start the blocking web server loop
	// this will exit when the handler gets fired and calls server.Close()
	server.Serve(l)
}

func getAccessToken(clientID, clientSecret, codeVerifier, authorizationCode, callbackURL string) (models.Token, error) {
	// set the url and form-encoded data for the POST to the access token endpoint
	url := "https://dev-y3bdlwyd.us.auth0.com/oauth/token"
	data := fmt.Sprintf(
		"grant_type=authorization_code&client_id=%s"+
			"&client_secret=%s"+
			"&code_verifier=%s"+
			"&code=%s"+
			"&redirect_uri=%s",
		clientID, clientSecret, codeVerifier, authorizationCode, callbackURL)
	payload := strings.NewReader(data)

	// create the request and execute it
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("HTTP error: %s", err)
		return models.Token{}, err
	}

	// process the response
	defer res.Body.Close()
	var responseData map[string]interface{}
	body, _ := ioutil.ReadAll(res.Body)

	// unmarshal the json into a string map
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Printf("JSON error: %s", err)
		return models.Token{}, err
	}

	// fmt.Println(responseData)
	// retrieve the access token out of the map, and return to caller
	accessToken := responseData["access_token"].(string)
	expiresAt := float64(time.Now().Unix()) + responseData["expires_in"].(float64)

	tokenStruct := models.Token{
		AccessToken:  accessToken,
		ExpiresIn:    responseData["expires_in"].(float64),
		ExpiresAt:    expiresAt,
		RefreshToken: responseData["refresh_token"].(string),
		IdToken:      responseData["id_token"].(string),
		TokenType:    responseData["token_type"].(string),
	}
	return tokenStruct, nil
}

func getUser(accessToken string) (models.User, error) {
	// set the url and form-encoded data for the POST to the access token endpoint
	url := "https://dev-y3bdlwyd.us.auth0.com/userinfo"

	// create the request and execute it
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("HTTP error: %s", err)
		return models.User{}, err
	}

	// process the response
	defer res.Body.Close()
	var responseData map[string]interface{}
	body, _ := ioutil.ReadAll(res.Body)

	// unmarshal the json into a string map
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Printf("JSON error: %s", err)
		return models.User{}, err
	}

	// fmt.Println(responseData)
	// retrieve the access token out of the map, and return to caller
	user := models.User{
		Name:  responseData["name"].(string),
		Email: responseData["email"].(string),
		Sub:   responseData["sub"].(string),
	}
	return user, nil
}

func RefreshToken(clientID, clientSecret string) {
	refreshToken := viper.Get("tokens.RefreshToken").(string)

	url := "https://dev-y3bdlwyd.us.auth0.com/oauth/token"
	data := fmt.Sprintf(
		"grant_type=refresh_token&client_id=%s"+
			"&client_secret=%s"+
			"&refresh_token=%s",
		clientID, clientSecret, refreshToken)
	payload := strings.NewReader(data)

	// create the request and execute it
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("HTTP error: %s", err)
		return
	}

	// process the response
	defer res.Body.Close()
	var responseData map[string]interface{}
	body, _ := ioutil.ReadAll(res.Body)

	// unmarshal the json into a string map
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Printf("JSON error: %s", err)
		return
	}

	// fmt.Println(responseData)
	// retrieve the access token out of the map, and return to caller
	accessToken := responseData["access_token"].(string)
	expiresAt := float64(time.Now().Unix()) + responseData["expires_in"].(float64)

	tokenStruct := models.Token{
		AccessToken:  accessToken,
		ExpiresIn:    responseData["expires_in"].(float64),
		ExpiresAt:    expiresAt,
		RefreshToken: refreshToken,
		IdToken:      responseData["id_token"].(string),
		TokenType:    responseData["token_type"].(string),
	}
	viper.Set("Tokens", tokenStruct)
	err = viper.WriteConfig()

	if err != nil {
		fmt.Println(err)
		fmt.Println("could not write config file")
		return
	}
}

func revoke(clientID, clientSecret string) {
	refreshToken := viper.Get("tokens.RefreshToken").(string)

	url := "https://dev-y3bdlwyd.us.auth0.com/oauth/revoke"
	data := fmt.Sprintf(
		"client_id=%s"+
			"&client_secret=%s"+
			"&token=%s",
		clientID, clientSecret, refreshToken)
	payload := strings.NewReader(data)

	// create the request and execute it
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("HTTP error: %s", err)
		return
	}

	// process the response
	defer res.Body.Close()
}

func Logout(clientID, clientSecret string) {

	revoke(clientID, clientSecret)

	url := "https://dev-y3bdlwyd.us.auth0.com/v2/logout?client_id" + clientID

	// create the request and execute it
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("content-type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("HTTP error: %s", err)
		return
	}

	// process the response
	defer res.Body.Close()

	viper.Set("Tokens", "")
	viper.Set("user", "")
	viper.WriteConfig()

	fmt.Println("Successfully logged out")
}

func cleanup(server *http.Server) {
	// we run this as a goroutine so that this function falls through and
	// the socket to the browser gets flushed/closed before the server goes away
	go server.Close()
}

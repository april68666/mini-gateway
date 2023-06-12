package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"mini-gateway/config"
	"mini-gateway/middleware"
	"mini-gateway/router/trie"
	"net/http"
	"strings"
)

const NAME = "jwt"

func init() {
	middleware.Register(NAME, Factory)
}

func Factory(c *config.Middleware) middleware.Middleware {
	secret := "12345678"
	if v, ok := c.Args["secret"]; ok {
		secret = v.(string)
	}

	t := trie.NewTrie[any]()
	if v, ok := c.Args["skipValidUrl"]; ok {
		ss := strings.Split(v.(string), ",")
		for _, s := range ss {
			t.Insert(s, nil)
		}
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return &jWt{
			next:   next,
			secret: secret,
			trie:   t,
		}
	}
}

type jWt struct {
	secret       string
	skipValidUrl string
	trie         *trie.Trie[any]
	next         http.RoundTripper
}

func (j *jWt) RoundTrip(req *http.Request) (*http.Response, error) {
	if _, _, b := j.trie.Search(req.URL.Path); b {
		return j.next.RoundTrip(req)
	}
	token := req.Header.Get("X-Auth-Token")
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
		}, nil
	}
	return j.next.RoundTrip(req)
}

//func createToken(userId int, secret string) (tokens string, err error) {
//	claim := jwt.MapClaims{
//		"id":  userId,
//		"nbf": time.Now().Unix(),
//		"iat": time.Now().Unix(),
//		"exp": time.Now().Unix() + 2592000,
//	}
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
//	tokens, err = token.SignedString([]byte(secret))
//	return
//}

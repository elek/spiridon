package bot

import (
	"net/http"
	"strings"
)

type Ntfy struct {
}

func (n *Ntfy) Send(destination string, msg string) error {
	req, _ := http.NewRequest("POST", "https://ntfy.sh/"+destination, strings.NewReader(msg))
	_, err := http.DefaultClient.Do(req)
	return err
}

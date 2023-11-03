package converter

type HttpPoster func(urlPath, filePath, token string) (string, error)

type Md2Html struct {
	ghToken string
	ghURL   string
	poster  HttpPoster
}

func NewMd2Html(ghToken, ghURL string, poster HttpPoster) Md2Html {
	return Md2Html{
		ghToken: ghToken,
		ghURL:   ghURL,
		poster:  poster,
	}
}

func (m Md2Html) Convert(path string) (string, error) {
	ghURL := m.ghURL + "/markdown/raw"
	return m.poster(ghURL, path, m.ghToken)
}

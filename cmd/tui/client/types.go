package client

type Tokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type loginEnvelope struct {
	Tokens Tokens `json:"tokens"`
}
